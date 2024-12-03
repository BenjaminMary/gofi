package sqlite

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"gofi/gofi/data/appdata"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
)

func ExportCSV(ctx context.Context, db *sql.DB, gofiID int, csvSeparator rune, csvDecimalDelimiter string, dateFormat string, dateSeparator string) (int, []byte) {
	/* take all data from the DB for a specific gofiID and put it in a csv file
	1. read database with gofiID
	2. write row by row in a csv (include headers)
	*/
	q := ` 
		SELECT fT.id, year, month, day, fT.mode,
			account, product, priceIntx100, category, 
			commentInt, commentString, checked, dateChecked, 
			IFNULL(lB.name, '')
		FROM financeTracker AS fT
			LEFT JOIN specificRecordsByMode AS sR ON fT.gofiID = sR.gofiID AND sR.idFinanceTracker = fT.id
			LEFT JOIN lenderBorrower AS lB ON lB.gofiID = sR.gofiID AND sR.idLenderBorrower = lB.id
		WHERE fT.gofiID = ?
			AND exported = 0
		ORDER BY fT.id
		LIMIT 10000;
	`
	rows, err := db.QueryContext(ctx, q, gofiID)
	if err != nil {
		fmt.Printf("error on SELECT financeTracker in ExportCSV, id: %v, err: %#v\n", gofiID, err)
	}
	file, _ := os.Create(appdata.SQLiteFilePath("gofi-" + strconv.Itoa(gofiID) + ".csv"))
	defer file.Close()
	w := csv.NewWriter(file)
	w.Comma = csvSeparator //french CSV file = ;
	defer w.Flush()

	var nbRows int = 0
	var row []string
	for rows.Next() {
		nbRows += 1
		if nbRows == 1 {
			//write csv headers
			row = []string{"𫝀é ꮖꭰ", "Date", "Mode",
				"Account", "Product", "PriceStr", "Category", "ThirdParty",
				"CommentInt", "CommentString", "Checked", "DateChecked", "Exported",
				"."} //keeping a column at the end will handle the LF and CRLF cases
			if err := w.Write(row); err != nil {
				fmt.Printf("row error 1: %v\n", row)
				log.Fatalln("error writing record to file", err)
			}
		}
		var ft appdata.FinanceTracker
		var successfull bool
		var unsuccessfullReason, lbName string
		if err := rows.Scan(
			&ft.ID, &ft.DateDetails.Year, &ft.DateDetails.Month, &ft.DateDetails.Day, &ft.Mode,
			&ft.Account, &ft.Product, &ft.PriceIntx100, &ft.Category,
			&ft.CommentInt, &ft.CommentString, &ft.Checked, &ft.DateChecked,
			&lbName,
		); err != nil {
			log.Fatal(err)
		}
		ft.FormPriceStr2Decimals = strings.Replace(ConvertPriceIntToStr(ft.PriceIntx100, true), ".", csvDecimalDelimiter, 1) //replace . to , for french CSV files
		ft.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, dateFormat, dateSeparator)
		if !successfull {
			ft.Date = "ERROR " + unsuccessfullReason
		}

		row = []string{strconv.Itoa(ft.ID), ft.Date, strconv.Itoa(ft.Mode),
			ft.Account, ft.Product, ft.FormPriceStr2Decimals, ft.Category, lbName,
			strconv.Itoa(ft.CommentInt), ft.CommentString, strconv.FormatBool(ft.Checked), ft.DateChecked, "true", "."}
		if err := w.Write(row); err != nil {
			fmt.Printf("row error 2: %v\n", row)
			log.Fatalln("error writing record to file", err)
		}
	}
	if nbRows == 0 {
		row = []string{"Rien à télécharger"}
		if err := w.Write(row); err != nil {
			fmt.Printf("row error 3: %v\n", row)
			log.Fatalln("error writing record to file", err)
		}
	}
	rows.Close()
	w.Flush() // write in the csv file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		fmt.Printf("ExportCSV err reseting the pointer to the start: %v\n", err)
	}
	fileData, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("ExportCSV err reading csv: %v\n", err)
	}
	q = ` 
		UPDATE financeTracker
		SET exported = 1
		WHERE gofiID = ?
			AND id IN (
				SELECT id
				FROM financeTracker
				WHERE gofiID = ?
					AND exported = 0
				ORDER BY id
				LIMIT 10000		
			);
	`
	_, err = db.ExecContext(ctx, q, gofiID, gofiID)
	if err != nil {
		fmt.Printf("error on UPDATE financeTracker with exported = 1, id: %v, err: %#v\n", gofiID, err)
	}
	return nbRows, fileData
}
func ExportCSVreset(ctx context.Context, db *sql.DB, gofiID int) {
	q := ` 
		UPDATE financeTracker
		SET exported = 0
		WHERE gofiID = ?
			AND exported = 1;
	`
	_, err := db.ExecContext(ctx, q, gofiID)
	if err != nil {
		fmt.Printf("error on UPDATE financeTracker with exported = 0, id: %v, err: %#v\n", gofiID, err)
	}
}

func ImportCSV(ctx context.Context, db *sql.DB,
	gofiID int, email string, csvSeparator rune, csvDecimalDelimiter string, dateFormat string, dateSeparator string, csvFile *multipart.FileHeader) (string, bool) {
	/* take all data from the csv and put it in the DB with a specific gofiID
	1. rows without ID are new ones (INSERT)
	2. rows with ID are existing ones (UPDATE)
	3. read csv (from line 2)
	4. write row by row in DB
	*/
	var stringList string
	var errorBool bool = false
	stringList += "traitement fichier pour: " + email + "\n"

	if csvFile.Size > 1000000 {
		stringList += "Fichier trop lourd: " + strconv.FormatInt(csvFile.Size, 10)
		stringList += " octets.\nLa limite actuelle est fixée à 1 000 000 octets par fichier.\nMerci de découper le fichier et faire plusieurs traitements."
		return stringList, true
	}
	file, err := csvFile.Open() // For read access.
	if err != nil {
		fmt.Printf("Unable to read input file: %v, error: %v", csvFile.Filename, err)
		stringList += "erreur d'ouverture du fichier csv, merci de vérifier le format."
		return stringList, true
	}
	defer file.Close() // this needs to be after the err check
	r := csv.NewReader(file)
	r.Comma = csvSeparator //french CSV file = ;
	rows, err := r.ReadAll()
	if err != nil {
		fmt.Printf("Unable to parse file as CSV for: %v, error: %v\n", csvFile.Filename, err)
		stringList += "erreur de lecture d'au moins 1 ligne dans le fichier csv, merci de vérifier le contenu et la structure du fichier."
		return stringList, true
	}

	var lb appdata.LendBorrow
	var ft appdata.FinanceTracker
	var iDlistInt []int
	var lineInfo, lineInfoTemp, unsuccessfullReason, controlEncoding, controlLastValidColumn, validControlEncodingUTF8, validControlEncodingUTF8withBOM string
	var successfull bool
	var flagErr, flagErrTemp int = 0, 0
	ft.GofiID = gofiID
	stringList += "𫝀é ꮖꭰ;Date;CommentInt;Checked;exported;NewID;Updated;\n"
	for index, row := range rows {
		if index == 0 { //control UTF-8 on headers
			totalRows := len(row)
			if totalRows != 14 {
				stringList =
					"IMPORTATION ANNULEE.\n" +
						"ERREUR sur le nombre de colonnes du fichier.\n\n" +
						"INFO: total " + strconv.Itoa(totalRows) + " colonnes sur un attendu de 14.\n" +
						"Un exemple de données d'import valide est disponible plus bas sur cette page."
				errorBool = true
				fmt.Printf("INFO: total %v colonnes sur un attendu de 14.\n", strconv.Itoa(totalRows))
				break //stop
			}
			controlEncoding = row[0]
			controlLastValidColumn = row[12]
			validControlEncodingUTF8 = "𫝀é ꮖꭰ"              //UTF-8
			validControlEncodingUTF8withBOM = "\ufeff𫝀é ꮖꭰ" //UTF-8 with BOM
			if (controlEncoding == validControlEncodingUTF8 || controlEncoding == validControlEncodingUTF8withBOM) &&
				controlLastValidColumn == "Exported" {
				continue //skip the row
			} else if controlLastValidColumn != "Exported" {
				fmt.Printf("totalRows: %#v\n", totalRows)
				fmt.Printf("controlEncoding: %#v\n", controlEncoding)
				stringList =
					"IMPORTATION ANNULEE.\n" +
						"ERREUR sur la dernière colonne du fichier.\n\n" +
						"INFO: 13eme colonne = 'Exported'\n" +
						"Un exemple de données d'import valide est disponible plus bas sur cette page."
				errorBool = true
				break //stop
			} else if !(controlEncoding == validControlEncodingUTF8 || controlEncoding == validControlEncodingUTF8withBOM) {
				fmt.Printf("totalRows: %#v\n", totalRows)
				fmt.Printf("controlEncoding: %#v\n", controlEncoding)
				stringList =
					"IMPORTATION ANNULEE.\n" +
						"ERREUR sur le format d'encodage du fichier.\n" +
						"Le système accepte uniquement du UTF-8 avec ou sans BOM.\n\n" +
						"INFO: des caractères spécifiques sont présents en en-tête de la 1ere colonne et doivent être gardés.\n" +
						"1ere colonne = '𫝀é ꮖꭰ'\n" +
						"Un exemple de données d'import valide est disponible plus bas sur cette page."
				errorBool = true
				break //stop
			}
		}
		lineInfo = ""
		ft.ID, err = strconv.Atoi(row[0])
		if err != nil { // Always check errors even if they should not happen.
			ft.ID = 0
			lineInfo += "INSERT;"
			// flagErr += 1 : not and ERROR, standard behaviour for INSERT
		} else {
			if ft.ID > 0 {
				lineInfo += "UPDATE " + row[0] + ";"
			} else if ft.ID < 0 {
				// DELETE is actually an UPDATE with empty data
				lineInfo += "DELETE" + row[0] + ";1999-12-31;;checked true;exported false;"
				ft.Date = "1999-12-31"
				ft.DateDetails.Year = 1999
				ft.DateDetails.Month = 12
				ft.DateDetails.Day = 31
				ft.Mode = 0
				ft.Account = "-"
				ft.Product = "DELETED LINE"
				ft.PriceIntx100 = 0
				ft.Category = "-"
				ft.CommentInt = 0
				ft.CommentString = ""
				ft.Checked = true //no need to validate a deleted row
				ft.DateChecked = "1999-12-31"
			} else if ft.ID == 0 {
				lineInfo += "INSERT;"
			}
		}

		if ft.ID >= 0 {
			ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, successfull, unsuccessfullReason = ConvertDateStrToInt(row[1], dateFormat, dateSeparator)
			if !successfull {
				lineInfo += "error " + unsuccessfullReason + ";;;;;false;"
				stringList += lineInfo + "\n"
				continue //skip this row because wrong date format
			}
			ft.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, "EN", "-")
			if !successfull {
				ft.Date = "ERROR " + unsuccessfullReason
			}


			ft.Mode, err = strconv.Atoi(row[2])
			if err != nil {
				ft.Mode = 0
			}
			ft.Account = row[3]
			ft.Product = row[4]
			ft.FormPriceStr2Decimals = row[5]
			ft.PriceIntx100 = ConvertPriceStrToInt(ft.FormPriceStr2Decimals, csvDecimalDelimiter)
			if ft.PriceIntx100 < 0 {
				ft.PriceDirection = "expense"
				ft.FormPriceStr2Decimals = ft.FormPriceStr2Decimals[1:] //remove the first "-"
			} else {
				ft.PriceDirection = "gain"
				// ft.FormPriceStr2Decimals = ft.FormPriceStr2Decimals
			}

			ft.Category = row[6]
			// ThirdParty = row[7] //handled bellow in lb var
			ft.CommentInt, err = strconv.Atoi(row[8])
			if err != nil {
				ft.CommentInt = 0
				lineInfo += "comment i 0;"
			} else {
				lineInfo += ";"
			}
			ft.CommentString = row[9]

			// Checked
			ft.Checked, err = strconv.ParseBool(row[10])
			if err != nil {
				ft.Checked = false
				lineInfo += "checked 0;"
			} else {
				lineInfo += ";"
			}

			// DateChecked
			ft.DateChecked = "9999-12-31"
			if len(row[11]) == 10 {
				yearInt, monthInt, dayInt, successfull, _ := ConvertDateStrToInt(row[11], dateFormat, dateSeparator)
				// fmt.Println("---------------")
				// fmt.Printf("ft.DateChecked: %v\n", ft.DateChecked)
				// fmt.Printf("yearInt %v, monthInt %v, dayInt %v, successfull %v, unsuccessfullReason %v\n", yearInt, monthInt, dayInt, successfull, unsuccessfullReason)
				if successfull {
					dateForDB, successfull, _ := ConvertDateIntToStr(yearInt, monthInt, dayInt, "EN", "-") //force YYYY-MM-DD inside DB
					//fmt.Printf("dateForDB %v, successfull %v, unsuccessfullReason %v\n", dateForDB, successfull, unsuccessfullReason)
					if successfull {
						ft.DateChecked = dateForDB
					}
				}
			}
			lineInfo += "Exported 0;"
		}

		if ft.ID < 0 { //DELETE part which is an UPDATE
			ft.ID = ft.ID * -1 //we keep the original positive ID, and send it to the standard UPDATE process
		}
		if ft.ID == 0 {
			// INSERT
			if ft.Mode == 0 {
				idFT, isErr, _, info := handleFTinsertCSV(ctx, &ft, csvSeparator, csvDecimalDelimiter, dateFormat, dateSeparator)
				if isErr {
					lineInfo += "error1;false;"
					fmt.Printf("error1: %v\n", info)
					flagErr += 1
					// return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, httpCode, info, "")
				} else {
					lineInfo += strconv.FormatInt(idFT, 10) + ";true;"
				}
			} else if ft.Mode >= 1 && ft.Mode <= 4 {
				lb.FT = ft
				lb.ModeInt = ft.Mode
				lb.CreateLenderBorrowerName = row[7]
				lb.Who = row[7]			
				lineInfoTemp, flagErrTemp = handleLBinsertCSV(ctx, &lb, csvSeparator, csvDecimalDelimiter, dateFormat, dateSeparator)
				lineInfo += lineInfoTemp
				flagErr += flagErrTemp
			} else {
				lineInfo += "error0;false;"
				fmt.Printf("error0, incorrect ft.Mode: %v\n", ft.Mode)
				flagErr += 1
			}
		} else if ft.ID > 0 {
			// UPDATE
			_, err = UpdateRowInFinanceTrackerFull(ctx, appdata.DB, &ft)
			if err != nil {
				lineInfo += "error3;false;"
				fmt.Printf("error3: %#v\n", err)
				flagErr += 1
			} else {
				if ft.Mode == 0 {
					iDlistInt = append(iDlistInt, ft.ID)
					lineInfo += ";true;"
				} else if ft.Mode >= 1 && ft.Mode <= 4 {
					lb.FT = ft
					lb.ModeInt = ft.Mode
					lb.CreateLenderBorrowerName = row[7]
					lb.Who = row[7]			
					lineInfoTemp, flagErrTemp = handleLBinsertCSV(ctx, &lb, csvSeparator, csvDecimalDelimiter, dateFormat, dateSeparator)
					lineInfo += lineInfoTemp
					flagErr += flagErrTemp
					lineInfo += ";true;"
				} else {
					iDlistInt = append(iDlistInt, ft.ID)
					lineInfo += "error4;false;"
					fmt.Printf("error4, incorrect ft.Mode: %v\n", ft.Mode)
					flagErr += 1
				}
			}
		}
		stringList += lineInfo + "\n"
	}
	b := DeleteSpecificRecordsByMode(ctx, appdata.DB, gofiID, &iDlistInt)
	if b {
		fmt.Println("DeleteSpecificRecordsByMode error in csv")
	}
	b = UpdateRowsInFinanceTrackerToMode0(ctx, appdata.DB, gofiID, &iDlistInt)
	if b {
		fmt.Println("UpdateRowsInFinanceTrackerToMode0 error in csv")
	}
	stringList = "erreurs rencontrées: " + strconv.Itoa(flagErr) + "\n" + stringList
	return stringList, errorBool
}

func handleLBinsertCSV(ctx context.Context, lb *appdata.LendBorrow,
	csvSeparator rune, csvDecimalDelimiter string, dateFormat string, dateSeparator string) (string, int) {
	// ft.Mode >= 1 && ft.Mode <= 4 {
	var lineInfo string = ""
	var flagErr int = 0
	isErr := InsertUpdateInLenderBorrower(ctx, appdata.DB, lb) // 1.
	if isErr {
		lineInfo += "handleLBinsertCSV-error1;false;"
		fmt.Println("handleLBinsertCSV-error1")
		flagErr += 1
	} else {
		var idFT int64
		var isErr bool
		var info string
		if lb.FT.ID == 0 {
			idFT, isErr, _, info = handleFTinsertCSV(ctx, &lb.FT, csvSeparator, csvDecimalDelimiter, dateFormat, dateSeparator) // 2.
		} else {
			idFT = int64(lb.FT.ID)
			isErr = false
		}
		if isErr {
			lineInfo += "handleLBinsertCSV-error2;false;"
			fmt.Printf("handleLBinsertCSV-error2: %v\n", info)
			flagErr += 1
		} else {
			lb.FT.ID = int(idFT)
			isErr = InsertInSpecificRecordsByMode(ctx, appdata.DB, lb) // 3.
			if isErr {
				lineInfo += "handleLBinsertCSV-error3;false;"
				fmt.Println("handleLBinsertCSV-error3")
				flagErr += 1
			} else {
				lineInfo += strconv.FormatInt(idFT, 10) + ";true;"
			}
		}
	}
	return lineInfo, flagErr
}

func handleFTinsertCSV(ctx context.Context, ft *appdata.FinanceTracker, 
	csvSeparator rune, csvDecimalDelimiter string, dateFormat string, dateSeparator string) (int64, bool, int, string) {
	if (ft.Mode == 1 || ft.Mode == 4) && ft.PriceDirection == "expense" {
		fmt.Println("handleFTinsert error4 invalid PriceDirection and Mode combinaison")
		return 0, true, 0, "invalid PriceDirection and Mode combinaison"
	}
	if (ft.Mode == 2 || ft.Mode == 3) && ft.PriceDirection == "gain" {
		fmt.Println("handleFTinsert error5 invalid PriceDirection and Mode combinaison")
		return 0, true, 0, "invalid PriceDirection and Mode combinaison"
	}
	if ft.PriceDirection == "" {
		fmt.Println("handleFTinsert error6 invalid PriceDirection")
		return 0, true, 0, "invalid PriceDirection"
	} else if ft.PriceDirection == "expense" {
		ft.FormPriceStr2Decimals = "-" + ft.FormPriceStr2Decimals
	}
	if ft.DateChecked == "" {
		ft.DateChecked = "9999-12-31"
	}
	ft.PriceIntx100 = ConvertPriceStrToInt(ft.FormPriceStr2Decimals, csvDecimalDelimiter) 
	idInserted, err := InsertRowInFinanceTracker(ctx, appdata.DB, ft)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("handleFTinsert error sqlite.InsertRowInFinanceTracker: %v\n", err.Error())
		return 0, true, 0, "server error"
	}
	return idInserted, false, 0, ""
}