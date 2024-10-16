CREATE TABLE notifications (
	id SERIAL PRIMARY KEY,
	user_id INT NOT NULL,
	type VARCHAR(50) NOT NULL,
	"from_user_id" INT NOT NULL,
	resource_id INT NOT NULL,
	is_read BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id),
	FOREIGN KEY ("from_user_id") REFERENCES users(id)
);


CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_from_user_id ON notifications("from_user_id");
CREATE INDEX idx_notifications_resource_id ON notifications(resource_id);