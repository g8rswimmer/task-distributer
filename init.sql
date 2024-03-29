CREATE TABLE IF NOT EXISTS SKILLS(
    SKILL VARCHAR(100) NOT NULL,
    DESCRIPTION TEXT NOT NULL,
    PRIMARY KEY(SKILL)
);

CREATE TABLE IF NOT EXISTS AGENTS(
    ID VARCHAR(10) NOT NULL,
    FIRSTNAME VARCHAR(100) NOT NULL,
    LASTNAME VARCHAR(100) NOT NULL,
    PRIMARY KEY(ID)
);

CREATE TABLE IF NOT EXISTS AGENTSKILLS(
    ID VARCHAR(10) NOT NULL,
    SKILL VARCHAR(100) REFERENCES SKILLS(SKILL),
    AGENT VARCHAR(10) REFERENCES AGENTS(ID),
    PRIMARY KEY(ID)
);

CREATE TABLE IF NOT EXISTS PRIORITIES(
    PRIORITY VARCHAR(100) NOT NULL,
    PRIORITY_LEVEL INT NOT NULL,
    PRIMARY KEY(PRIORITY) 
);

CREATE TABLE IF NOT EXISTS TASKS(
    ID VARCHAR(100) NOT NULL,
    CREATEDATE TIMESTAMP NOT NULL,
    NAME TEXT NOT NULL,
    SKILLS TEXT[],
    PRIORITY VARCHAR(100) REFERENCES PRIORITIES(PRIORITY),
    STATUS VARCHAR(100) NOT NULL,
    COMPLETEDATE TIMESTAMP,  
    AGENT VARCHAR(10) REFERENCES AGENTS(ID) 
);

DO $$
BEGIN
IF NOT EXISTS(SELECT * FROM SKILLS) THEN
INSERT INTO SKILLS 
    (SKILL, DESCRIPTION)
VALUES 
    ('skill1', 'This is a great skill to have'),
    ('skill2', 'This is a awesome skill to have'),
    ('skill3', 'This is a cool skill to have');
END IF;	

IF NOT EXISTS(SELECT * FROM AGENTS) THEN
INSERT INTO AGENTS 
    (ID, FIRSTNAME, LASTNAME)
VALUES
    ('1000', 'Bighead', 'Burton'),
    ('1001', 'Ovaltine', 'Jenkins'),    
    ('1002', 'Ground', 'Control'),
    ('1003', 'Jazz', 'Hands');    
END IF;	

IF NOT EXISTS(SELECT * FROM AGENTSKILLS) THEN
INSERT INTO AGENTSKILLS 
    (ID, SKILL, AGENT)
VALUES
    ('2000', 'skill1', '1000'),
    ('2001', 'skill2', '1001'),    
    ('2002', 'skill3', '1001'),
    ('2003', 'skill3', '1002'),    
    ('2004', 'skill1', '1003'),    
    ('2005', 'skill3', '1003');    
END IF;	
IF NOT EXISTS(SELECT * FROM PRIORITIES) THEN
INSERT INTO PRIORITIES 
    (PRIORITY, PRIORITY_LEVEL)
VALUES
    ('low', 0),
    ('high', 1);    
END IF;	
END
$$

