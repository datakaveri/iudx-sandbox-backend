CREATE TABLE IF NOT EXISTS public.gallery
(
    "repoName"       TEXT,
    "description"    TEXT,
    "githubUrl"      TEXT           NOT NULL,
    "createdAt"      TIMESTAMPTZ    NOT NULL    DEFAULT NOW(),
    "updatedAt"      TIMESTAMPTZ    NOT NULL    DEFAULT NOW(),
    PRIMARY KEY ("repoName")
);
