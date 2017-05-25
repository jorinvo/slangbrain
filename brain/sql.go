package brain

var (
	addPhraseStmt = `
		INSERT INTO phrases (chatid, phrase, explanation)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	addStudyStmt = `
		INSERT INTO studies (phraseid, studymode) VALUES ($1, $2)
	`

	getModeStmt = `
		SELECT mode FROM chats WHERE chatid = $1
	`

	setModeStmt = `
		INSERT INTO chats (chatid, mode) VALUES ($1, $2)
		ON CONFLICT (chatid)
		DO UPDATE
		SET mode = $2
		WHERE chats.chatid = $1
	`

	getStudyStmt = `
		with
		grouped as (
			select sum(score) as sumscore, timestamp, studies.id as id, studymode, phrase, explanation
			from studies
			join phrases
			on phraseid = phrases.id
			where chatid = $1
			group by phraseid, studymode, studies.id, phrase, explanation
			order by timestamp
		),
		filtered as (
			select id, studymode, phrase, explanation
			from grouped
			where now() - timestamp >= (power(2, sumscore + 1) || ' hours')::interval
		),
		total as (
			select count(1)
			from filtered
		)

		select *
		from filtered
		cross join total
		limit 1
	`

	scoreStmt = `
		insert into studies (phraseid, studymode, score)
		with
		grouped as (
			select sum(score) as sumscore, timestamp, phraseid, studymode
			from studies
			join phrases
			on phraseid = phrases.id
			where chatid = $1
			group by phraseid, studymode, studies.id
			order by timestamp
		),
		filtered as (
			select phraseid, studymode
			from grouped
			where now() - timestamp >= (power(2, sumscore + 1) || ' hours')::interval
		),
		score as (
			select ($2)::integer
		)

		select *
		from filtered
		cross join score
		limit 1
	`

	countStudiesStmt = `
		TODO
	`
)
