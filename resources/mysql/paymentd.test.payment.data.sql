START TRANSACTION;

INSERT INTO `payment_method` (`id`, `project_id`, `provider_id`, `method_key`, `created`, `created_by`) 
	VALUES (1, 1, 1, 'test', UTC_TIMESTAMP(), 'test');

COMMIT;
