CREATE TABLE IF NOT EXISTS public.gallery
(
    "repoName"      TEXT,
    "description"   TEXT,
    "githubUrl"     TEXT            NOT NULL,
    "createdAt"     TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    "updatedAt"     TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    PRIMARY KEY ("repoName")
);

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
