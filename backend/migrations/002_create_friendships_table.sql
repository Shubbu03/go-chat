-- Create friendships table
CREATE TABLE IF NOT EXISTS friendships (
    id SERIAL PRIMARY KEY,
    requester_id INTEGER NOT NULL,
    addressee_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (addressee_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Ensure valid status values
    CONSTRAINT friendships_status_check CHECK (status IN ('pending', 'accepted', 'rejected', 'blocked')),
    
    -- Prevent duplicate friendship requests between same users
    CONSTRAINT friendships_unique_pair UNIQUE (requester_id, addressee_id),
    
    -- Prevent self-friendship
    CONSTRAINT friendships_no_self_friend CHECK (requester_id != addressee_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_friendships_requester_id ON friendships(requester_id);
CREATE INDEX IF NOT EXISTS idx_friendships_addressee_id ON friendships(addressee_id);
CREATE INDEX IF NOT EXISTS idx_friendships_status ON friendships(status);
CREATE INDEX IF NOT EXISTS idx_friendships_user_pair ON friendships(requester_id, addressee_id);
CREATE INDEX IF NOT EXISTS idx_friendships_deleted_at ON friendships(deleted_at);

-- Index for finding all friendships for a user (both directions)
CREATE INDEX IF NOT EXISTS idx_friendships_user_relationships ON friendships(requester_id, addressee_id, status);

-- Add trigger to update updated_at timestamp
CREATE TRIGGER update_friendships_updated_at BEFORE UPDATE ON friendships 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 