START TRANSACTION;

INSERT INTO `payment_method` (`id`, `project_id`, `provider_id`, `method_key`, `created`, `created_by`) 
	VALUES (1, 1, 1, 'test', UTC_TIMESTAMP(), 'test');
INSERT INTO `payment_method_status` (`payment_method_id`, `timestamp`, `created_by`, `status`)
	VALUES (1, 1, 'test', 'active');
	
COMMIT;