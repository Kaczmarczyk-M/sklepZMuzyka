DROP TABLE IF EXISTS product;
CREATE TABLE product (
  productid         INT unsigned AUTO_INCREMENT NOT NULL ,
  title      VARCHAR(128) NOT NULL,
  artist     VARCHAR(255) NOT NULL,
  price      DECIMAL(5,2) NOT NULL,
  PRIMARY KEY (`productid`)
);

INSERT INTO product 
  (title, artist, price) 
VALUES 
  ('Favourite Worst Nightmare', 'Arctic Monkeys', 39.99),
  ('The Wall', 'Pink Floyd', 29.99),
  ('Led Zeppelin III', 'Led Zeppelin', 45.00),
  ('Sticky Fingers', 'The Rolling Stones', 39.99);
