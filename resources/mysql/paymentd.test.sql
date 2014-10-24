-- MySQL Script generated by MySQL Workbench
-- Fri 24 Oct 2014 01:40:22 PM CEST
-- Model: New Model    Version: 1.0
SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL,ALLOW_INVALID_DATES';

-- -----------------------------------------------------
-- Schema fritzpay_payment_test
-- -----------------------------------------------------
DROP SCHEMA IF EXISTS `fritzpay_payment_test` ;
CREATE SCHEMA IF NOT EXISTS `fritzpay_payment_test` DEFAULT CHARACTER SET utf8mb4 ;
-- -----------------------------------------------------
-- Schema fritzpay_principal_test
-- -----------------------------------------------------
DROP SCHEMA IF EXISTS `fritzpay_principal_test` ;
CREATE SCHEMA IF NOT EXISTS `fritzpay_principal_test` DEFAULT CHARACTER SET utf8mb4 ;
USE `fritzpay_payment_test` ;

-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`config`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`config` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`config` (
  `name` VARCHAR(64) NOT NULL,
  `last_change` DATETIME NOT NULL,
  `value` TEXT NULL,
  PRIMARY KEY (`name`, `last_change`))
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`provider`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`provider` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`provider` (
  `id` INT UNSIGNED NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `name_UNIQUE` (`name` ASC))
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_principal_test`.`principal`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_principal_test`.`principal` ;

CREATE TABLE IF NOT EXISTS `fritzpay_principal_test`.`principal` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created` DATETIME NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `name_UNIQUE` (`name` ASC))
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_principal_test`.`project`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_principal_test`.`project` ;

CREATE TABLE IF NOT EXISTS `fritzpay_principal_test`.`project` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `principal_id` INT UNSIGNED NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `created` DATETIME NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `project_name` (`principal_id` ASC, `name` ASC),
  CONSTRAINT `fk_project_principal_id`
    FOREIGN KEY (`principal_id`)
    REFERENCES `fritzpay_principal_test`.`principal` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`payment_method`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`payment_method` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`payment_method` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` INT UNSIGNED NOT NULL,
  `provider_id` INT UNSIGNED NOT NULL,
  `method_key` VARCHAR(64) NOT NULL,
  `created` DATETIME NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_payment_method_project_id_idx` (`project_id` ASC),
  INDEX `fk_payment_method_provider_id_idx` (`provider_id` ASC),
  UNIQUE INDEX `method_key` (`project_id` ASC, `provider_id` ASC, `method_key` ASC),
  CONSTRAINT `fk_payment_method_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `fritzpay_principal_test`.`project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,
  CONSTRAINT `fk_payment_method_provider_id`
    FOREIGN KEY (`provider_id`)
    REFERENCES `fritzpay_payment_test`.`provider` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`currency`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`currency` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`currency` (
  `code_iso_4217` VARCHAR(3) NOT NULL,
  PRIMARY KEY (`code_iso_4217`))
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`payment_method_status`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`payment_method_status` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`payment_method_status` (
  `payment_method_id` BIGINT UNSIGNED NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  PRIMARY KEY (`payment_method_id`, `timestamp`),
  CONSTRAINT `fk_payment_method_status_payment_method_id`
    FOREIGN KEY (`payment_method_id`)
    REFERENCES `fritzpay_payment_test`.`payment_method` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`payment_method_metadata`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`payment_method_metadata` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`payment_method_metadata` (
  `payment_method_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `value` TEXT NOT NULL,
  PRIMARY KEY (`payment_method_id`, `name`, `timestamp`),
  CONSTRAINT `fk_principal_metadata_payment_method_id`
    FOREIGN KEY (`payment_method_id`)
    REFERENCES `fritzpay_payment_test`.`payment_method` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`payment`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`payment` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`payment` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` INT UNSIGNED NOT NULL,
  `created` DATETIME NOT NULL,
  `ident` VARCHAR(175) NOT NULL,
  `amount` INT NOT NULL,
  `subunits` TINYINT(4) UNSIGNED NOT NULL,
  `currency` VARCHAR(3) NOT NULL,
  `callback_url` TEXT NULL,
  `return_url` TEXT NULL,
  PRIMARY KEY (`id`),
  INDEX `created` (`created` ASC),
  UNIQUE INDEX `ident` (`project_id` ASC, `ident` ASC),
  INDEX `fk_payment_currency_idx` (`currency` ASC),
  UNIQUE INDEX `payment_id` (`id` ASC, `project_id` ASC),
  CONSTRAINT `fk_payment_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `fritzpay_principal_test`.`project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,
  CONSTRAINT `fk_payment_currency`
    FOREIGN KEY (`currency`)
    REFERENCES `fritzpay_payment_test`.`currency` (`code_iso_4217`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`payment_config`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`payment_config` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`payment_config` (
  `project_id` INT UNSIGNED NOT NULL,
  `payment_id` BIGINT UNSIGNED NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `payment_method_id` BIGINT UNSIGNED NULL,
  `country` VARCHAR(2) NULL,
  `locale` VARCHAR(5) NULL,
  PRIMARY KEY (`project_id`, `payment_id`, `timestamp`),
  INDEX `fk_payment_config_payment_method_id_idx` (`payment_method_id` ASC),
  INDEX `fk_payment_config_payment_id_idx` (`payment_id` ASC),
  CONSTRAINT `fk_payment_config_payment_id`
    FOREIGN KEY (`payment_id`)
    REFERENCES `fritzpay_payment_test`.`payment` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,
  CONSTRAINT `fk_payment_config_payment_method_id`
    FOREIGN KEY (`payment_method_id`)
    REFERENCES `fritzpay_payment_test`.`payment_method` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`payment_metadata`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`payment_metadata` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`payment_metadata` (
  `project_id` INT UNSIGNED NOT NULL,
  `payment_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(125) NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `value` TEXT NULL,
  PRIMARY KEY (`project_id`, `payment_id`, `name`, `timestamp`),
  INDEX `fk_payment_metadata_payment_id_idx` (`payment_id` ASC),
  CONSTRAINT `fk_payment_metadata_payment_id`
    FOREIGN KEY (`payment_id`)
    REFERENCES `fritzpay_payment_test`.`payment` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,
  CONSTRAINT `fk_payment_metadata_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `fritzpay_principal_test`.`project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`payment_token`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`payment_token` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`payment_token` (
  `token` VARCHAR(64) NOT NULL,
  `created` DATETIME NOT NULL,
  `project_id` INT UNSIGNED NOT NULL,
  `payment_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`token`),
  INDEX `created` (`created` ASC),
  INDEX `fk_payment_token_payment_id_idx` (`payment_id` ASC),
  INDEX `fk_payment_token_project_id_idx` (`project_id` ASC),
  CONSTRAINT `fk_payment_token_payment_id`
    FOREIGN KEY (`payment_id`)
    REFERENCES `fritzpay_payment_test`.`payment` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,
  CONSTRAINT `fk_payment_token_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `fritzpay_principal_test`.`project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_payment_test`.`payment_transaction`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_payment_test`.`payment_transaction` ;

CREATE TABLE IF NOT EXISTS `fritzpay_payment_test`.`payment_transaction` (
  `project_id` INT UNSIGNED NOT NULL,
  `payment_id` BIGINT UNSIGNED NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `amount` INT NOT NULL,
  `subunits` TINYINT(4) UNSIGNED NOT NULL,
  `currency` VARCHAR(3) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `comment` TEXT NULL,
  PRIMARY KEY (`project_id`, `payment_id`, `timestamp`),
  INDEX `status` (`status` ASC),
  INDEX `fk_payment_transaction_currency_idx` (`currency` ASC),
  INDEX `fk_payment_transaction_payment_id_idx` (`payment_id` ASC),
  CONSTRAINT `fk_payment_transaction_payment_id`
    FOREIGN KEY (`payment_id`)
    REFERENCES `fritzpay_payment_test`.`payment` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,
  CONSTRAINT `fk_payment_transaction_currency`
    FOREIGN KEY (`currency`)
    REFERENCES `fritzpay_payment_test`.`currency` (`code_iso_4217`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,
  CONSTRAINT `fk_payment_transaction_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `fritzpay_principal_test`.`project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;

USE `fritzpay_principal_test` ;

-- -----------------------------------------------------
-- Table `fritzpay_principal_test`.`principal_metadata`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_principal_test`.`principal_metadata` ;

CREATE TABLE IF NOT EXISTS `fritzpay_principal_test`.`principal_metadata` (
  `principal_id` INT UNSIGNED NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `value` TEXT NOT NULL,
  PRIMARY KEY (`principal_id`, `name`, `timestamp`),
  CONSTRAINT `fk_principal_metadata_principal_id`
    FOREIGN KEY (`principal_id`)
    REFERENCES `fritzpay_principal_test`.`principal` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_principal_test`.`project_metadata`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_principal_test`.`project_metadata` ;

CREATE TABLE IF NOT EXISTS `fritzpay_principal_test`.`project_metadata` (
  `project_id` INT UNSIGNED NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `value` TEXT NOT NULL,
  PRIMARY KEY (`project_id`, `name`, `timestamp`),
  CONSTRAINT `fk_project_metadata_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `fritzpay_principal_test`.`project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `fritzpay_principal_test`.`project_key`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `fritzpay_principal_test`.`project_key` ;

CREATE TABLE IF NOT EXISTS `fritzpay_principal_test`.`project_key` (
  `key` VARCHAR(64) NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `project_id` INT UNSIGNED NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `secret` TEXT NOT NULL,
  `active` TINYINT(1) NOT NULL,
  PRIMARY KEY (`key`, `timestamp`),
  INDEX `fk_project_key_project_id_idx` (`project_id` ASC),
  CONSTRAINT `fk_project_key_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `fritzpay_principal_test`.`project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;

SET SQL_MODE = '';
GRANT USAGE ON *.* TO paymentd;
 DROP USER paymentd;
SET SQL_MODE='TRADITIONAL,ALLOW_INVALID_DATES';
CREATE USER 'paymentd';

GRANT SELECT, INSERT ON TABLE fritzpay_payment_test.* TO 'paymentd';
GRANT SELECT, INSERT ON TABLE fritzpay_principal_test.* TO 'paymentd';

SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;

-- -----------------------------------------------------
-- Data for table `fritzpay_payment_test`.`provider`
-- -----------------------------------------------------
START TRANSACTION;
USE `fritzpay_payment_test`;
INSERT INTO `fritzpay_payment_test`.`provider` (`id`, `name`) VALUES (1, 'fritzpay');

COMMIT;

