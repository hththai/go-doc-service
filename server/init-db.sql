-- Grant CREATE DATABASE privilege to appuser
GRANT CREATE ON *.* TO 'appuser'@'%';
FLUSH PRIVILEGES;
