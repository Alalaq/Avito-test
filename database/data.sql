drop table if exists segment_history;
drop table if exists user_segments;
drop table if exists segments;
drop table if exists users;

CREATE TABLE segments (
                          id INT AUTO_INCREMENT PRIMARY KEY,
                          slug VARCHAR(255) NOT NULL,
                          auto_add BOOLEAN DEFAULT false,
                          auto_pct INT DEFAULT 0,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE users (
                       id INT AUTO_INCREMENT PRIMARY KEY,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_segments (
                               user_id INT,
                               segment_id INT,
                               expires_at DATETIME,
                               PRIMARY KEY (user_id, segment_id),
                               FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
                               FOREIGN KEY (segment_id) REFERENCES segments (id) ON DELETE CASCADE
);

CREATE TABLE segment_history (
                                 id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
                                 user_id INT NOT NULL,
                                 segment_id INT NOT NULL,
                                 operation VARCHAR(20) NOT NULL,
                                 timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                 FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                                 FOREIGN KEY (segment_id) REFERENCES segments(id) ON DELETE CASCADE
);

