package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gofi/gofi/data/appdata"
	"time"
)

func CheckUserLogin(ctx context.Context, db *sql.DB, user *appdata.User) (int, string, error) {
	q := ` 
		SELECT gofiID
		FROM user
		WHERE email = ?
			AND pwHash = ?;
	`
	var gofiID int = 0
	err := db.QueryRowContext(ctx, q, user.Email, user.PwHash).Scan(&gofiID)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("CheckUserLogin error no row returned, user.Email: %v\n", user.Email)
		return 0, "error no row returned", err
	case err != nil:
		fmt.Printf("CheckUserLogin error: %v\n", err)
		return 0, "error querying DB", err
	}
	if gofiID > 0 {
		_, err := db.ExecContext(ctx, `
			UPDATE user 
			SET numberOfRequests = numberOfRequests + 1,
				idleTimeout = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', idleDateModifier)),
				absoluteTimeout = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', absoluteDateModifier)),
				lastLoginTime = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now')), 
				lastActivityTime = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now')), 
				sessionID = ?, lastActivityIPaddress = ?, lastActivityUserAgent = ?, lastActivityAcceptLanguage = ?
			WHERE gofiID = ?;
			`,
			user.SessionID, user.LastActivityIPaddress, user.LastActivityUserAgent, user.LastActivityAcceptLanguage,
			gofiID,
		)
		if err != nil {
			return gofiID, "error on UPDATE after login", err
		}
	}
	return gofiID, "", nil
}

func UpdateSessionID(ctx context.Context, db *sql.DB, gofiID int, sessionID string) (string, error) {
	// fmt.Printf("in UpdateSessionID: %v, new: %v", gofiID, sessionID)
	_, err := db.ExecContext(ctx, `
		UPDATE user 
		SET idleTimeout = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', idleDateModifier)),
			lastActivityTime = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now')), 
			sessionID = ?
		WHERE gofiID = ?;
		`,
		// --, lastActivityIPaddress = ?, lastActivityUserAgent = ?, lastActivityAcceptLanguage = ?
		sessionID, //user.LastActivityIPaddress, user.LastActivityUserAgent, user.LastActivityAcceptLanguage,
		gofiID,
	)
	if err != nil {
		return "error on UPDATE for UpdateSessionID", err
	}
	return "", nil
}

func Logout(ctx context.Context, db *sql.DB, gofiID int) (bool, string, error) {
	if gofiID > 0 {
		_, err := db.ExecContext(ctx, `
			UPDATE user 
			SET numberOfRequests = numberOfRequests + 1,
				idleTimeout = '1999-12-31T00:01:01Z',
				absoluteTimeout = '1999-12-31T00:01:01Z',
				lastActivityTime = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now')), 
				sessionID = NULL
			WHERE gofiID = ?;
			`,
			gofiID,
		)
		if err != nil {
			fmt.Printf("error on UPDATE in logout: %v\n", err)
			return false, "error on UPDATE in logout", err
		}
	}
	return true, "", nil
}

func GetGofiID(ctx context.Context, db *sql.DB, sessionID string, 
	newActivityUserAgent string, newActivityAcceptLanguage string, newActivityIPaddress string) (int, string, string, error) {
	q := ` 
		SELECT u.gofiID, u.email, u.idleTimeout, u.absoluteTimeout, 
			u.lastActivityUserAgent, u.lastActivityAcceptLanguage, u.lastActivityIPaddress,
			strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now')) AS currentTimeUTC,
			p.paramJSONstringData
		FROM user AS u
			LEFT JOIN param AS p ON u.gofiID = p.gofiID AND p.paramName = 'forceNewLoginOnIPchange'
		WHERE u.sessionID = ?
			AND u.sessionID IS NOT NULL
			AND u.sessionID NOT LIKE 'logged-out-%';
	`
	var gofiID int = 0
	var email, idleTimeout, absoluteTimeout, lastActivityUserAgent, lastActivityAcceptLanguage, lastActivityIPaddress, currentTimeUTC string
	var forceNewLoginOnIPchange string
	err := db.QueryRowContext(ctx, q, sessionID).Scan(&gofiID, &email, &idleTimeout, &absoluteTimeout, 
		&lastActivityUserAgent, &lastActivityAcceptLanguage, &lastActivityIPaddress, &currentTimeUTC,
		&forceNewLoginOnIPchange)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("GetGofiID error no row returned, sessionID: %v\n", sessionID)
		return 0, "", "error no row returned", err
	case err != nil:
		fmt.Printf("GetGofiID error: %v\n", err)
		return 0, "", "error querying DB", err
	}

	timeCurrentTimeUTC, err := time.Parse(time.RFC3339, currentTimeUTC)
	// fmt.Printf("timeCurrentTimeUTC: %v\n", timeCurrentTimeUTC)
	if err != nil {
		return gofiID, "", "error parsing currentTimeUTC, force new login 1", err
	}

	timeAbsoluteTimeout, err := time.Parse(time.RFC3339, absoluteTimeout)
	// fmt.Printf("timeAbsoluteTimeout: %v\n", timeAbsoluteTimeout)
	if err != nil {
		return gofiID, "", "error parsing absoluteTimeout, force new login 2", err
	}
	differenceAbsolute := timeCurrentTimeUTC.Sub(timeAbsoluteTimeout)
	// fmt.Printf("differenceAbsolute: %v\n", differenceAbsolute)
	if differenceAbsolute > 0 {
		return gofiID, "", "absoluteTimeout, force new login 3", errors.New("absolute-timeout")
	}

	timeIdleTimeout, err := time.Parse(time.RFC3339, idleTimeout)
	// fmt.Printf("timeIdleTimeout: %v\n", timeIdleTimeout)
	if err != nil {
		return gofiID, "", "error parsing idleTimeout, force new login 4", err
	}

	/*fmt.Printf("newActivityUserAgent: %v, lastActivityUserAgent: %v," + 
		"newActivityAcceptLanguage: %v, lastActivityAcceptLanguage: %v," +
		"newActivityIPaddress: %v, lastActivityIPaddress: %v\n", 
		newActivityUserAgent, lastActivityUserAgent, newActivityAcceptLanguage, lastActivityAcceptLanguage, newActivityIPaddress, lastActivityIPaddress)
	*/
	if newActivityUserAgent != lastActivityUserAgent {
		return gofiID, "", "userAgentChange, force new login 5", errors.New("user-agent-change")
	}
	if newActivityAcceptLanguage != lastActivityAcceptLanguage {
		return gofiID, "", "acceptLanguageChange, force new login 6", errors.New("accept-language-change")
	}
	if forceNewLoginOnIPchange == "1" {
		if newActivityIPaddress != lastActivityIPaddress {
			return gofiID, "", "IPaddressChange, force new login 7", errors.New("IP-address-change")
		}
	}

	differenceIdle := timeCurrentTimeUTC.Sub(timeIdleTimeout)
	// fmt.Printf("differenceIdle: %v\n", differenceIdle)
	if differenceIdle > 0 {
		return gofiID, email, "idleTimeout, change cookie", nil
	}

	if gofiID > 0 {
		return gofiID, email, "", nil
	} else {
		return 0, "", "error no gofiID found from sessionID cookie", errors.New("no-gofiID-found")
	}
}

func CreateUser(ctx context.Context, db *sql.DB, user appdata.User) (int64, string, error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO user (email, pwHash, dateCreated)
		VALUES (?,?,?);
		`,
		user.Email, user.PwHash, user.DateCreated,
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

func DeleteUser(ctx context.Context, db *sql.DB, gofiID int) (string, error) {
	_, err := db.ExecContext(ctx, `
		DELETE FROM user 
		WHERE gofiID = ?;
		`,
		gofiID,
	)
	if err != nil {
		return "error deleting user row in DB", err
	}
	_, err = db.ExecContext(ctx, `
		DELETE FROM param 
		WHERE gofiID = ?;
		`,
		gofiID,
	)
	if err != nil {
		return "error deleting param rows in DB", err
	}
	return "", nil
}
