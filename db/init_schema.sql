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
	"id"                VARCHAR(200)    			NOT NULL,
	"accessPolicy"      VARCHAR(30)                 NOT NULL,
	"description"       TEXT                        NOT NULL,
	"domain"			VARCHAR(100),
	"icon"              VARCHAR(300),
	"iudxResourceAPIs"  VARCHAR[],
	"label"             VARCHAR(100)                NOT NULL,
	"name"              VARCHAR(100)                NOT NULL,
	"provider"          json,
	"repositoryURL"     VARCHAR(300),
	"tags"              VARCHAR[],
	"type"              VARCHAR[],
	"unique_id"         VARCHAR(200) PRIMARY KEY 	not null,
	"resources"			INT							not NULL,
	"instance"			VARCHAR(100)				not NULL
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
	foreign key ("resourceGroup") 		references dataset("unique_id")
);

create table referenceResources (
	"id"				varchar(200) 	primary key not null,
	"name"				varchar(100),
	"description"		text,
	"additionalInfoURL" VARCHAR(300),
	"datasetID"			varchar(300)				not null,
	foreign key ("datasetID") references dataset("unique_id")
);


