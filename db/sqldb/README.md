# Placeholders

## 1. Static Placeholders
For static placeholders, we use `?` and then convert them to each DBMS with conversion function.

### MySQL
Since MySQL uses `?` for placeholders, no conversion is required.

### PostgreSQL
`?` -> `$n` where n = 1, 2, 3, ...

### MS SQL
`?` -> `@n` where n = 1, 2, 3, ...

### Oracle
`?` -> `:n` where n = 1, 2, 3, ...

## 2. Dynamic Placeholders
A dynamic placeholder is a notation used to represent a variable number of placeholders.
It uses a symbol `??`, which can be converted—via a conversion function—into forms like `?, ?, ..., ?` or `$k, $k+1, ..., $n`, depending on the DBMS.


## Prepared Statements
Since we store raw SQL statements in the banks after conversion for static placeholders only, they can be used as prepared statements if they don't contain dynamic placeholders. 