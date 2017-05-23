INSERT INTO chats (chatid) VALUES (1373046609441564);

INSERT INTO phrases (phrase, explanation, chatid) VALUES
  ('Hallo', 'Hello', 1373046609441564),
  ('Wie geht es dir?', 'How are you?', 1373046609441564),
  ('Ich habe hunger.', 'I am hungry.', 1373046609441564),
  ('Hilfe!', 'Help!', 1373046609441564);

INSERT INTO studies (phraseid, studymode, score, timestamp) VALUES
  /* ready to study */
  (1, 1, 0, DateTime('Now', '-1 Day')),
  /* just added. need to wait. */
  (2, 1, 0, DateTime('Now')),
  /* not ready to study because of sumscore */
  (3, 1,  0, DateTime('Now', '-6 Days')),
  (3, 1,  1, DateTime('Now', '-5 Days')),
  (3, 1, -1, DateTime('Now', '-5 Days')),
  (3, 1,  1, DateTime('Now', '-4 Days')),
  (3, 1,  1, DateTime('Now', '-2 Days')),
  (3, 1,  1, DateTime('Now', '-1 Day')),
  /* ready to study */
  (4, 1, 0, DateTime('Now', '-5 Days')),
  (4, 1, 1, DateTime('Now', '-3 Days'));
