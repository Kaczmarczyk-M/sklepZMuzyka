DROP TABLE IF EXISTS customer;

CREATE TABLE customer (
  customerid         INT unsigned AUTO_INCREMENT NOT NULL ,
  email      VARCHAR(100) NOT NULL,
  pass     VARCHAR(100) NOT NULL,
  PRIMARY KEY (`customerid`)
);

INSERT INTO customer
    (email,pass)
VALUES 
  ('m01kaczmarczyk@gmail.com', 'qwerty'),
  ('sample@gmail.com', 'haslo123');
