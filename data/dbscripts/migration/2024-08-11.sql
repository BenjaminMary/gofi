--DROP TABLE IF EXISTS modes;
CREATE TABLE IF NOT EXISTS modes (
    mode INTEGER NOT NULL, 
    info TEXT NOT NULL
);
INSERT INTO modes
VALUES 
    (0, '+- standard'),
    (1, '+ emprunt'),
    (2, '- pret'),
    (3, '- remboursement emprunt'),
    (4, '+ remboursement pret')
;


--DROP TABLE IF EXISTS lenderBorrower;
CREATE TABLE IF NOT EXISTS lenderBorrower (
    id INTEGER PRIMARY KEY AUTOINCREMENT, 
    gofiID INTEGER NOT NULL,
    name TEXT NOT NULL,
    isActive INTEGER DEFAULT 1
);

--DROP TABLE IF EXISTS specificRecordsByMode;
CREATE TABLE IF NOT EXISTS specificRecordsByMode (
    id INTEGER PRIMARY KEY AUTOINCREMENT, 
    gofiID INTEGER NOT NULL,
    mode INTEGER NOT NULL,
    idFinanceTracker INTEGER NOT NULL,
    idLenderBorrower INTEGER DEFAULT 0

    -- parentIdIfRefund INTEGER DEFAULT 0,
);


ALTER TABLE financeTracker 
ADD COLUMN mode INTEGER DEFAULT 0;
