CREATE TABLE teams (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) REFERENCES users(id),
    status VARCHAR(10) CHECK (status IN ('OPEN', 'MERGED')) DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT NOW(),
    merged_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE pr_reviewers (
    pr_id VARCHAR(255) REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id VARCHAR(255) REFERENCES users(id),
    assigned_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (pr_id, user_id)
);