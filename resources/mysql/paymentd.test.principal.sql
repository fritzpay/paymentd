-- MySQL Script generated by MySQL Workbench
-- Wed 29 Oct 2014 06:45:48 PM CET
-- Model: New Model    Version: 1.0
SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL,ALLOW_INVALID_DATES';

-- -----------------------------------------------------
-- Schema fritzpay_payment
-- -----------------------------------------------------
-- -----------------------------------------------------
-- Schema fritzpay_principal
-- -----------------------------------------------------

-- -----------------------------------------------------
-- Table `principal`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `principal` ;

CREATE TABLE IF NOT EXISTS `principal` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created` DATETIME NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `name_UNIQUE` (`name` ASC))
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `principal_status`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `principal_status` ;

CREATE TABLE IF NOT EXISTS `principal_status` (
  `principal_id` INT UNSIGNED NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  PRIMARY KEY (`principal_id`, `timestamp`),
  CONSTRAINT `fk_principal_status_principal_id`
    FOREIGN KEY (`principal_id`)
    REFERENCES `principal` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `project`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `project` ;

CREATE TABLE IF NOT EXISTS `project` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `principal_id` INT UNSIGNED NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `created` DATETIME NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `project_name` (`principal_id` ASC, `name` ASC),
  CONSTRAINT `fk_project_principal_id`
    FOREIGN KEY (`principal_id`)
    REFERENCES `principal` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `principal_metadata`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `principal_metadata` ;

CREATE TABLE IF NOT EXISTS `principal_metadata` (
  `principal_id` INT UNSIGNED NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `value` TEXT NOT NULL,
  PRIMARY KEY (`principal_id`, `name`, `timestamp`),
  CONSTRAINT `fk_principal_metadata_principal_id`
    FOREIGN KEY (`principal_id`)
    REFERENCES `principal` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `project_metadata`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `project_metadata` ;

CREATE TABLE IF NOT EXISTS `project_metadata` (
  `project_id` INT UNSIGNED NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `timestamp` BIGINT UNSIGNED NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `value` TEXT NOT NULL,
  PRIMARY KEY (`project_id`, `name`, `timestamp`),
  CONSTRAINT `fk_project_metadata_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `project_key`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `project_key` ;

CREATE TABLE IF NOT EXISTS `project_key` (
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
    REFERENCES `project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `project_config`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `project_config` ;

CREATE TABLE IF NOT EXISTS `project_config` (
  `project_id` INT UNSIGNED NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `web_url` TEXT NULL,
  `callback_url` TEXT NULL,
  `callback_api_version` VARCHAR(32) NULL,
  `callback_project_key` VARCHAR(64) NULL,
  `return_url` TEXT NULL,
  PRIMARY KEY (`project_id`, `timestamp`),
  INDEX `fk_project_config_project_key_idx` (`callback_project_key` ASC),
  CONSTRAINT `fk_project_config_callback_project_key`
    FOREIGN KEY (`callback_project_key`)
    REFERENCES `project_key` (`key`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,
  CONSTRAINT `fk_project_config_project_id`
    FOREIGN KEY (`project_id`)
    REFERENCES `project` (`id`)
    ON DELETE RESTRICT
    ON UPDATE CASCADE)
ENGINE = InnoDB;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
