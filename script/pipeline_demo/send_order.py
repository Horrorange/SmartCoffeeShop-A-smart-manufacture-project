import os
import time
import psycopg

def env(k, d):
    v = os.environ.get(k)
    return v if v else d

def get_conn():
    host = env('PG_HOST','localhost')
    port = env('PG_PORT','5432')
    db = env('PG_DB','smartshop')
    user = env('PG_USER','postgres')
    pw = env('PG_PASS','885658')
    return psycopg.connect(host=host, port=port, dbname=db, user=user, password=pw)

def ensure_schema(cur):
    cur.execute('''
    CREATE TABLE IF NOT EXISTS orders (
      id BIGSERIAL PRIMARY KEY,
      coffee_type VARCHAR(32) NOT NULL,
      bool_ice BOOLEAN NOT NULL,
      table_num INT NOT NULL,
      status VARCHAR(16) NOT NULL DEFAULT 'pending',
      create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      finish_time TIMESTAMPTZ,
      error_msg TEXT
    );
    ''')

def insert_orders(cur):
    orders = [
        ('LATTE', True, 7),
        ('AMERICANO', False, 3),
        ('ESPRESSO', False, 5),
        ('MOCHA', True, 8),
    ]
    for ct, ice, table in orders:
        cur.execute("INSERT INTO orders(coffee_type, bool_ice, table_num, status, create_time) VALUES(%s,%s,%s,'pending',NOW())", (ct, ice, table))

def list_orders(cur):
    cur.execute("SELECT id, coffee_type, bool_ice, table_num, status, COALESCE(error_msg,'') FROM orders ORDER BY id")
    return cur.fetchall()

def main():
    with get_conn() as conn:
        with conn.cursor() as cur:
            ensure_schema(cur)
            insert_orders(cur)
        conn.commit()
    print('Inserted pending orders.')
    # poll
    for _ in range(30):
        with get_conn() as conn:
            with conn.cursor() as cur:
                rows = list_orders(cur)
                for r in rows:
                    print(r)
        time.sleep(2)

if __name__ == '__main__':
    main()