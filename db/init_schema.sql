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
    "url"           TEXT,
    "name"          TEXT,
    "buildId"       UUID,
    "status"        TEXT            NOT NULL,
    "lastUsed"      TIMESTAMPTZ     NOT NULL    DEFAULT NOW(), 
    "createdAt"     TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    "updatedAt"     TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    UNIQUE ("buildId"),
    PRIMARY KEY ("notebookId")
);

CREATE TABLE IF NOT EXISTS public.buildlog
(
    "id"            SERIAL,
    "buildId"       UUID            NOT NULL,
    "phase"         TEXT,
    "message"       TEXT,
    "token"         TEXT,
    "progress"      JSONB,
    "imageName"     TEXT,
    "url"           TEXT,
    "createdAt"     TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    "updatedAt"     TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    PRIMARY KEY ("id")
    CONSTRAINT "notebookFK"
        FOREIGN KEY ("buildId") 
        REFERENCES "notebook" ("buildId"),
);