DROP TABLE IF EXISTS orders;
CREATE TABLE if NOT EXISTS orders (
  orderid INT unsigned AUTO_INCREMENT NOT NULL,
  productid INT unsigned not null,
  customerid INT unsigned not null,
  timeunix timestamp not null,
  PRIMARY KEY (`orderid`),
  FOREIGN KEY (`productid`) REFERENCES product(productid),
  FOREIGN KEY (`customerid`) REFERENCES customer(customerid)
);

INSERT INTO orders
  (productid, customerid,timeunix) 
VALUES 
  (3,1,current_timestamp),
  (2,1,current_timestamp);
