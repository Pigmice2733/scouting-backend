ALTER TABLE reports DROP CONSTRAINT reports_reporter_fkey;
ALTER TABLE reports ADD FOREIGN KEY(reporter) REFERENCES users(username) ON UPDATE CASCADE