CREATE DATABASE secrets;
CREATE USER secrets WITH ENCRYPTED PASSWORD 'secrets';
GRANT ALL PRIVILEGES ON DATABASE secrets TO secrets;
ALTER ROLE secrets SUPERUSER;
