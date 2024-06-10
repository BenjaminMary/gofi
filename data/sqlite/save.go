package sqlite

import (
	"context"
	"database/sql"
	"gofi/gofi/data/appdata"
)

// r.Get("/", api.SaveRead)      // GET /api/save
// r.Post("/", api.SaveCreate)   // POST /api/save
// r.Put("/", api.SaveEdit)      // PUT /api/save
// r.Delete("/", api.SaveDelete) // DELETE /api/save

func SaveSelect(ctx context.Context, db *sql.DB) ([]appdata.SaveBackup, string, error) {
	saveBackupList := []appdata.SaveBackup{}
	q := `
		SELECT id, date, extID, extFileName, checkpoint, tested
		FROM backupSave
		ORDER BY id DESC;
	`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return saveBackupList, "select query error", err
	}
	for rows.Next() {
		saveBackup := appdata.SaveBackup{}
		if err := rows.Scan(&saveBackup.ID, &saveBackup.Date, &saveBackup.ExtID, &saveBackup.ExtFileName, &saveBackup.Checkpoint, &saveBackup.Tested); err != nil {
			return saveBackupList, "row scan error", err
		}
		saveBackupList = append(saveBackupList, saveBackup)
	}
	return saveBackupList, "", nil
}

func SaveCreate(ctx context.Context, db *sql.DB, saveBackup *appdata.SaveBackup) (int64, string, error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO backupSave (date, extID, extFileName, checkpoint, tested)
		VALUES (?,?,?,?,?);
		`,
		saveBackup.Date, saveBackup.ExtID, saveBackup.ExtFileName, saveBackup.Checkpoint, saveBackup.Tested,
	)
	if err != nil {
		return 0, "error inserting row in DB", err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, "error to get last inserted id in DB", err
	}
	return id, "", nil
}

// func SaveUpdate(ctx context.Context, db *sql.DB) (string, error) {
// 	_, err := db.ExecContext(ctx, `
// 		UPDATE backupSave
// 		SET checkpoint = ?, tested = ?
// 		WHERE id = ?
// 			AND extID = ?;
// 		`,
// 		sessionID, //user.LastActivityIPaddress, user.LastActivityUserAgent, user.LastActivityAcceptLanguage,
// 		gofiID,
// 	)
// 	if err != nil {
// 		return "SaveUpdate error", err
// 	}
// 	return "", nil
// }

func SaveDelete(ctx context.Context, db *sql.DB, id int) (string, error) {
	_, err := db.ExecContext(ctx, `
		DELETE FROM backupSave 
		WHERE id = ?;
		`,
		id,
	)
	if err != nil {
		return "error deleting backupSave row in DB", err
	}
	return "", nil
}

func SaveDeleteKeepX(ctx context.Context, db *sql.DB, keepX int) ([]string, string, error) {
	var driveIDlist []string
	driveIDlist = append(driveIDlist, "")
	q := `
		SELECT MAX(id) AS maxID
		FROM backupSave;
	`
	var maxID int = 0
	err := db.QueryRowContext(ctx, q).Scan(&maxID)
	switch {
	case err == sql.ErrNoRows:
		return driveIDlist, "error no row returned", err
	case err != nil:
		return driveIDlist, "error querying DB", err
	}
	maxID = maxID - keepX + 1
	if maxID > 0 {
		q := ` 
			SELECT extID
			FROM backupSave
			WHERE id < ?;
		`
		rows, _ := db.QueryContext(ctx, q, maxID)
		var driveID string
		for rows.Next() {
			if err := rows.Scan(&driveID); err != nil {
				return driveIDlist, "error querying SaveDeleteKeepX row in DB", err
			}
			driveIDlist = append(driveIDlist, driveID)
		}
		rows.Close()

		_, err := db.ExecContext(ctx, `
			DELETE FROM backupSave 
			WHERE id < ?;
			`,
			maxID,
		)
		if err != nil {
			return driveIDlist, "error deleting backupSave row in DB", err
		}
	}
	return driveIDlist, "", nil
}
