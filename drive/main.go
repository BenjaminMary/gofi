package drive

import (
        "fmt"
        "log"
        "os"

	"bytes"
        "net/http"
	"io"

        "encoding/json"

        "golang.org/x/oauth2"
        //"google.golang.org/api/drive/v3"
)

type DriveFileMetaDataList struct {
        Files []DriveFileMetaData `json:"files"`
}
type DriveFileMetaData struct {
	DriveFileID string `json:"id,omitempty"` // != driveId : which is populated for items in shared drives 
        Name string `json:"name"`
        Trashed bool `json:"trashed"`
        // ModifiedTime string `json:"modifiedTime"`
}

func UploadWithDrivePostRequestAPI(fileToUpload string) DriveFileMetaData {
	conf := GoogleAuth()
	client := conf.Client(oauth2.NoContext)
	
        file, err := os.ReadFile(fileToUpload)
        if err != nil {
           log.Fatalln(err)
        }

        url := "https://www.googleapis.com/upload/drive/v3/files?uploadType=media"
        //fmt.Println("URL: ", url)
        req, err := http.NewRequest("POST", url, bytes.NewBuffer(file))

        resp, err := client.Do(req)
        if err != nil {
                log.Fatal(err)
        }
        defer resp.Body.Close()

        //fmt.Println("response Status: ", resp.Status)
        // fmt.Println("response Headers:", resp.Header)
        body, _ := io.ReadAll(resp.Body)
        //fmt.Println("-----------------------------\nresponse Body POST:\n", string(body))

        var driveFileMetaData DriveFileMetaData
        json.Unmarshal(body, &driveFileMetaData)

        //fmt.Printf("-----------------------------\n json: %#v\n", driveFileMetaData)
        //fmt.Printf("-----------------------------\n json id: %v\n", driveFileMetaData.DriveFileID)

        return driveFileMetaData
}

func ListFileInDrive() DriveFileMetaDataList {
        conf := GoogleAuth()
        client := conf.Client(oauth2.NoContext)

        // ne retourne rien car compte de service : requestURL :="https://www.googleapis.com/drive/v3/drives"
        // |'me' in owners| = |%27me%27%20in%20owners| : permet de ne lister que les fichiers dans le compte de service
        requestURL := "https://www.googleapis.com/drive/v3/files?q=%27me%27%20in%20owners"

        req, err := http.NewRequest(http.MethodGet, requestURL, nil)
        if err != nil {
                log.Fatal(err)
        }
        resp, err := client.Do(req)
        if err != nil {
                log.Fatal(err)
        }
        bytesR, err := io.ReadAll(resp.Body)
        if err != nil {
                log.Fatal(err)
        }
        //fmt.Println("-----------------------------\nresponse Body LIST:\n", string(bytesR))

        var driveFileMetaDataList DriveFileMetaDataList
        json.Unmarshal(bytesR, &driveFileMetaDataList)

        return driveFileMetaDataList
}

func DeleteFileInDrive(driveID string) {
        conf := GoogleAuth()
        client := conf.Client(oauth2.NoContext)

        requestURL := "https://www.googleapis.com/drive/v3/files/" + driveID

        req, err := http.NewRequest("DELETE", requestURL, nil)
        if err != nil {
                log.Fatal(err)
        }
        resp, err := client.Do(req)
        if err != nil {
                log.Fatal(err)
        }
        _, err = io.ReadAll(resp.Body) // bytesR, err :=
        if err != nil {
                log.Fatal(err)
        }
        //fmt.Println("URL: ", requestURL)
        //fmt.Println("response Status: ", resp.Status)
        //fmt.Println("-----------------------------\nresponse Body DELETE:\n", string(bytesR))
}


func UpdateMetaDataDriveFile(dfmd DriveFileMetaData) {
	conf := GoogleAuth()
	client := conf.Client(oauth2.NoContext)
	
        // PATCH request
        url := "https://www.googleapis.com/drive/v3/files/" + dfmd.DriveFileID
        dfmd.DriveFileID = "" // remove the id for the JSON sent with omitempty

        //fmt.Println("URL: ", url)

        jsonB, err := json.Marshal(dfmd)
        req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonB))

        resp, err := client.Do(req)
        if err != nil {
                log.Fatal(err)
        }
        defer resp.Body.Close()

        //fmt.Println("response Status: ", resp.Status)
        // fmt.Println("response Headers:", resp.Header)
        //body, _ := io.ReadAll(resp.Body)
        //fmt.Println("-----------------------------\nresponse Body PATCH:\n", string(body))
}

func GetFileInDrive(driveID string, pathAndFileName string) {
        conf := GoogleAuth()
        client := conf.Client(oauth2.NoContext)

        // to download add : ?alt=media 
        // requestURL := "https://www.googleapis.com/drive/v3/files/" + driveID + "?fields=id,name,kind,mimeType,trashed,createdTime,modifiedTime,shared"
        requestURL := "https://www.googleapis.com/drive/v3/files/" + driveID + "?alt=media"

        req, err := http.NewRequest(http.MethodGet, requestURL, nil)
        if err != nil {
                log.Fatal(err)
        }
        resp, err := client.Do(req)
        if err != nil {
                log.Fatal(err)
        }
        defer resp.Body.Close()

        output, err := os.Create(pathAndFileName)
        if err != nil {
            fmt.Println("Error while creating file")
            log.Fatal(err)
        }
        defer output.Close()

	_, err = io.Copy(output, resp.Body) // n, err :
	if err != nil {
	        fmt.Println("Error while downloading")
                log.Fatal(err)
	}
        //fmt.Println(n, "bytes downloaded.")

        return
}