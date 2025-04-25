ALTER TABLE messages DROP CONSTRAINT IF EXISTS fk_receiver;

ALTER TABLE messages ADD CONSTRAINT fk_receiver_org 
    FOREIGN KEY (receiver_organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
