START TRANSACTION;

INSERT INTO `principal` (`id`, `created`, `created_by`, `name`) 
	VALUES (1, UTC_TIMESTAMP(), 'test', 'testprincipal');
INSERT INTO `project` (`id`, `principal_id`, `name`, `created`, `created_by`) 
	VALUES (1, 1, 'testproject', UTC_TIMESTAMP(), 'test');
INSERT INTO `project_key` (`key`, `timestamp`, `project_id`, `created_by`, `secret`, `active`)
	VALUES ('testkey', UTC_TIMESTAMP, 1, 'test', 'abcdef', 1);

COMMIT;
