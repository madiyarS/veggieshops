-- Seed initial data for development (run after 001_create_tables)
-- Categories
INSERT INTO categories (name, description, "order", is_active) VALUES
  ('Овощи', 'Свежие овощи', 1, true),
  ('Фрукты', 'Сезонные фрукты', 2, true),
  ('Зелень', 'Свежая зелень', 3, true),
  ('Ягоды', 'Ягоды', 4, true);
