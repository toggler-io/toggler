-- default deployment environment
ALTER TABLE "pilots"
    DROP COLUMN "env_id";

DELETE
FROM "deployment_environments"
WHERE "id" = '77f19375-0745-41b8-ad24-959008ca66ac';
