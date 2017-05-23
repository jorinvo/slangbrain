package brain

var (
	addPhraseStmt = `
		INSERT INTO phrases (chatid, phrase, explanation) VALUES ($1, $2, $3)
	`

	addStudyStmt = `
		INSERT INTO studies (phraseid, studymode) VALUES ($1, $2)
	`

	getModeStmt = `
		SELECT mode FROM chats WHERE chatid = $1
	`

	setModeStmt = `
		REPLACE INTO chats (chatid, mode) VALUES ($1, $2)
	`

	getStudyStmt = `
		WITH tmp AS (
			SELECT *
			FROM (
				SELECT SUM(score) AS sumscore, studies.id as id, studymode, phrase, explanation, timestamp
				FROM studies
				JOIN phrases
				ON phraseid = phrases.id
				WHERE chatid = $1
				GROUP BY phraseid, studymode
				ORDER BY timestamp
			)
			WHERE (julianday('now') - julianday(timestamp)) >= (2 << sumscore + 1) / 24.0
		)
		SELECT id, studymode, phrase, explanation, total
		FROM tmp
		JOIN (
			SELECT COUNT(1) AS total FROM tmp
		)
		LIMIT 1
	`

	scoreStmt = `
		INSERT INTO studies (phraseid, studymode, score)
		SELECT *
		FROM (
			SELECT phraseid, studymode
			FROM (
				SELECT SUM(score) AS sumscore, timestamp, phraseid, studymode
				FROM studies
				JOIN phrases
				ON phraseid = phrases.id
				WHERE chatid = 1373046609441564
				GROUP BY phraseid, studymode
				ORDER BY timestamp
			)
			WHERE (julianday('now') - julianday(timestamp)) >= (2 << sumscore + 1) / 24.0
			LIMIT 1
		)
		JOIN (SELECT 20)
	`
)
