-- default deployment environment
INSERT INTO "deployment_environments" ("id", "name")
VALUES ('77f19375-0745-41b8-ad24-959008ca66ac', 'default');

ALTER TABLE "pilots"
    ADD COLUMN "env_id" UUID NOT NULL DEFAULT '77f19375-0745-41b8-ad24-959008ca66ac';

ALTER TABLE "pilots"
    ALTER COLUMN "env_id" DROP DEFAULT;

DELETE
FROM "deployment_environments" AS d
WHERE NOT EXISTS(
        SELECT
        FROM "pilots" AS p
        WHERE p.env_id = d.id
    );