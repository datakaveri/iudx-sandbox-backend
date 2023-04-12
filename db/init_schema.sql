CREATE TABLE IF NOT EXISTS public.notebook
(
    "notebookId"    UUID            NOT NULL,
    "userId"		INTEGER,
    "spawnerId"     INTEGER,
    "name"          TEXT,
    "repoName"      TEXT,
    "buildId"       UUID,
    "phase"         TEXT,
    "message"       TEXT,
    "token"         TEXT,
    "imageName"     TEXT,
    "url"           TEXT,
    "lastUsed"      TIMESTAMPTZ     NOT NULL    DEFAULT NOW(), 
    "createdAt"     TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    "updatedAt"     TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    UNIQUE ("buildId"),
    PRIMARY KEY ("notebookId"),
    CONSTRAINT "userFK"
        FOREIGN KEY ("userId") 
        REFERENCES "users" ("id"),
    CONSTRAINT "spawnerFK"
        FOREIGN KEY ("spawnerId")
        REFERENCES "spawners" ("id")
);

create table dataset (
	"id"                VARCHAR(200)    PRIMARY KEY NOT NULL,
	"accessPolicy"      VARCHAR(30)                 NOT NULL,
	"createdAt"         TIMESTAMP                   NOT NULL,
	"description"       TEXT                        NOT NULL,
	"icon"              VARCHAR(300),
	"instance"          VARCHAR(30),
	"itemCreatedAt"     TIMESTAMP,
	"itemStatus"        VARCHAR(30),
	"iudxResourceAPIs"  VARCHAR[],
	"label"             VARCHAR(100)                NOT NULL,
	"location"          json,
	"name"              VARCHAR(100)                NOT NULL,
	"provider"          json,
--	"referenceResources" jsonb,
	"repositoryURL"     VARCHAR(300)                NOT NULL,
	"resourceServer"    VARCHAR(300)                NOT NULL,
	"resourceType"      VARCHAR(30),
	"resources"         INT,
	"schema"            VARCHAR(300),
	"tags"              VARCHAR[],
	"type"              VARCHAR[],
	"unique_id"         VARCHAR(200) 				not null,
	"updatedAt"         TIMESTAMP,
	"views"             INT
);

create table resource (
	"id" 				VARCHAR(200) 	primary key not null,
	"createdAt" 		timestamp 					not null,
	"dataset" 			varchar(100),
	"description" 		text 						not null,
	"downloadURL" 		varchar(300) 				not null,
	"icon" 				varchar(300) 				not null,
	"instance" 			varchar(20) 				not null,
	"itemCreatedAt" 	timestamp 					not null,
	"itemStatus" 		varchar(30),
	"label" 			varchar(100),
	"name" 				varchar(100),
	"provider" 			varchar(100),
	"resourceGroup" 	varchar(300),
	"tags" 				VARCHAR[],
	"type" 				VARCHAR[],
	"updatedAt" 		timestamp,
	foreign key ("resourceGroup") 		references dataset("id")
);
