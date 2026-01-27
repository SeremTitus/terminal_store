INSERT INTO products (name, price, stock)
SELECT v.name, v.price, v.stock
FROM (
  VALUES
    ('Coffee Beans', 14.99, 50),
    ('Tea Sampler', 9.50, 40),
    ('Ceramic Mug', 12.00, 25)
) AS v(name, price, stock)
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.name = v.name);

INSERT INTO customers (name, phone)
SELECT v.name, v.phone
FROM (
  VALUES
    ('Ava Carter', '555-0101'),
    ('Miles Nguyen', '555-0134')
) AS v(name, phone)
WHERE NOT EXISTS (SELECT 1 FROM customers c WHERE c.name = v.name);

DO $$
DECLARE
  ava_id INT;
  miles_id INT;
  coffee_id INT;
  tea_id INT;
  mug_id INT;
BEGIN
  SELECT id INTO ava_id FROM customers WHERE name = 'Ava Carter';
  SELECT id INTO miles_id FROM customers WHERE name = 'Miles Nguyen';
  SELECT id INTO coffee_id FROM products WHERE name = 'Coffee Beans';
  SELECT id INTO tea_id FROM products WHERE name = 'Tea Sampler';
  SELECT id INTO mug_id FROM products WHERE name = 'Ceramic Mug';

  IF (SELECT COUNT(*) FROM orders) = 0 THEN
    INSERT INTO orders (customer_id)
    VALUES (ava_id), (miles_id);

    INSERT INTO order_items (order_id, product_id, qty, price_each)
    VALUES
      ((SELECT id FROM orders WHERE customer_id = ava_id ORDER BY id LIMIT 1), coffee_id, 2, 14.99),
      ((SELECT id FROM orders WHERE customer_id = ava_id ORDER BY id LIMIT 1), mug_id, 1, 12.00),
      ((SELECT id FROM orders WHERE customer_id = miles_id ORDER BY id LIMIT 1), tea_id, 3, 9.50);
  END IF;
END $$;
