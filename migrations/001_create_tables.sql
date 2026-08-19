-- Migration: create_tables
-- Created at: 2025-11-01 15:44:01

CREATE TABLE IF NOT EXISTS habits (
    id BIGSERIAL PRIMARY KEY,             
    title VARCHAR(100) NOT NULL,          
    description TEXT,                    
    color VARCHAR(7),                       
    icon VARCHAR(50),                      
    frequency VARCHAR(20) DEFAULT 'daily',  
    target_days INTEGER DEFAULT 7,         
    is_active BOOLEAN DEFAULT TRUE,     
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_habits_is_active ON habits(is_active);

CREATE TABLE IF NOT EXISTS habit_logs (
    id BIGSERIAL PRIMARY KEY,         
    habit_id BIGINT NOT NULL,      
    date DATE NOT NULL,        
    notes TEXT,                  
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_habit 
        FOREIGN KEY (habit_id) 
        REFERENCES habits(id) 
        ON DELETE CASCADE,                 
    
    CONSTRAINT unique_habit_date 
        UNIQUE(habit_id, date)
);

CREATE INDEX IF NOT EXISTS idx_habit_logs_habit_id ON habit_logs(habit_id);
CREATE INDEX IF NOT EXISTS idx_habit_logs_date ON habit_logs(date);
CREATE INDEX IF NOT EXISTS idx_habit_logs_habit_date ON habit_logs(habit_id, date);

-- автообновление даты
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_habits_updated_at ON habits;
CREATE TRIGGER update_habits_updated_at 
    BEFORE UPDATE ON habits 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
