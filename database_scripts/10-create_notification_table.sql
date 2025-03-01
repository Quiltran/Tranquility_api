CREATE TABLE notification (user_id int primary key, endpoint text, p256dh text, auth text);
ALTER TABLE notification ADD CONSTRAINT fk_notification_user_id FOREIGN KEY (user_id) REFERENCES auth(id);