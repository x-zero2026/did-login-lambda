-- X-Zero Platform Database Schema
-- Run this in your Supabase SQL editor

-- Users table
CREATE TABLE IF NOT EXISTS users (
    did VARCHAR(66) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    project_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_name VARCHAR(255) NOT NULL,
    creator_did VARCHAR(66) NOT NULL REFERENCES users(did) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_projects_creator ON projects(creator_did);

-- User-Project relationship table
CREATE TABLE IF NOT EXISTS user_projects (
    user_did VARCHAR(66) NOT NULL REFERENCES users(did) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'member')),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_did, project_id)
);

CREATE INDEX IF NOT EXISTS idx_user_projects_user ON user_projects(user_did, joined_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_projects_project ON user_projects(project_id);

-- Apps table
CREATE TABLE IF NOT EXISTS apps (
    app_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name VARCHAR(255) NOT NULL,
    app_description TEXT,
    emoji VARCHAR(10) NOT NULL,
    url VARCHAR(500) NOT NULL,
    is_global BOOLEAN DEFAULT FALSE,
    created_by_did VARCHAR(66) NOT NULL REFERENCES users(did) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_apps_global ON apps(is_global);
CREATE INDEX IF NOT EXISTS idx_apps_creator ON apps(created_by_did);

-- App-Project relationship table (for non-global apps)
CREATE TABLE IF NOT EXISTS app_projects (
    app_id UUID NOT NULL REFERENCES apps(app_id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id, project_id)
);

CREATE INDEX IF NOT EXISTS idx_app_projects_project ON app_projects(project_id);

-- Project-App settings table (for closing apps in specific projects)
CREATE TABLE IF NOT EXISTS project_app_settings (
    project_id UUID NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    app_id UUID NOT NULL REFERENCES apps(app_id) ON DELETE CASCADE,
    is_closed BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, app_id)
);

CREATE INDEX IF NOT EXISTS idx_project_app_settings_project ON project_app_settings(project_id);

-- Update updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_apps_updated_at BEFORE UPDATE ON apps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_project_app_settings_updated_at BEFORE UPDATE ON project_app_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
