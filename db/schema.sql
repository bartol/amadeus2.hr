CREATE TABLE IF NOT EXISTS grupe (
    sifra INT PRIMARY KEY,
    naziv TEXT
);

CREATE TABLE IF NOT EXISTS proizvodi (
    sifra INT PRIMARY KEY,
    grupa INT,
    FOREIGN KEY (grupa) REFERENCES grupe(sifra),
    naziv TEXT,
    kolicina INT,
    nabavna_cijena DECIMAL(9, 2),
    marza DECIMAL(9, 2),
    cijena DECIMAL(9, 2),
    rabat DECIMAL(9, 2),

    amadeus2hr CHAR(1),
    pioneerhr CHAR(1),
    njuskalohr CHAR(1),

    web_cijena DECIMAL(9, 2),
    web_cijena_s_popustom DECIMAL(9, 2)
);

-- CREATE INDEX amadeus2hr_idx ON proizvodi(amadeus2hr) WHERE amadeus2hr IS NOT NULL;
-- CREATE INDEX pioneerhr_idx ON proizvodi(pioneerhr) WHERE pioneerhr IS NOT NULL;
-- CREATE INDEX njuskalohr_idx ON proizvodi(njuskalohr) WHERE njuskalohr IS NOT NULL;
