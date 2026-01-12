BEGIN;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users 
ADD CONSTRAINT users_role_check 
CHECK (role IN ('user', 'owner', 'moderator'));

UPDATE users SET role = 'owner' WHERE role = 'admin';

COMMIT;
