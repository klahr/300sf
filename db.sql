CREATE TABLE `300sf`.`resources` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `type` VARCHAR(32) NOT NULL,
  `mimetype` VARCHAR(64) NOT NULL,
  `filename` VARCHAR(256) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE);

CREATE TABLE `300sf`.`users` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL DEFAULT NOW(),
  `uuid` VARCHAR(32) NOT NULL,
  `display_name` VARCHAR(128) NOT NULL,
  `email` VARCHAR(254) NOT NULL,
  `password` VARCHAR(64) NOT NULL,
  `password_salt` VARCHAR(64) NOT NULL,
  `permissions` INT NOT NULL,
  `icon_id` INT NULL,
  `notifications` INT NOT NULL,
  `deleted_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  UNIQUE INDEX `email_UNIQUE` (`email` ASC) VISIBLE,
  UNIQUE INDEX `password_UNIQUE` (`password` ASC) VISIBLE,
  UNIQUE INDEX `password_salt_UNIQUE` (`password_salt` ASC) VISIBLE,
  INDEX `index6` (`uuid` ASC) VISIBLE,
  INDEX `fk_users_1_idx` (`icon_id` ASC) VISIBLE,
  CONSTRAINT `fk_users_1`
    FOREIGN KEY (`icon_id`)
    REFERENCES `300sf`.`resources` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`api_keys` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL DEFAULT NOW(),
  `uuid` VARCHAR(32) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `prefix` VARCHAR(7) NOT NULL,
  `hash` VARCHAR(64) NOT NULL,
  `scopes` INT NOT NULL,
  `soft_limit` INT NOT NULL,
  `hard_limit` INT NOT NULL,
  `user_id` INT NOT NULL,
  `deactivated_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  UNIQUE INDEX `prefix_UNIQUE` (`prefix` ASC) VISIBLE,
  INDEX `index4` (`uuid` ASC) VISIBLE,
  INDEX `fk_api_keys_2_idx` (`user_id` ASC) VISIBLE,
  CONSTRAINT `fk_api_keys_2`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

ALTER TABLE `300sf`.`users` 
ADD COLUMN `confirmed_at` DATETIME NULL AFTER `notifications`;

ALTER TABLE `300sf`.`users` 
ADD COLUMN `last_seen_at` DATETIME NULL AFTER `notifications`,
ADD COLUMN `deactivated_at` DATETIME NULL AFTER `confirmed_at`;

ALTER TABLE `300sf`.`users` 
ADD COLUMN `accepted_terms_at` DATETIME NULL AFTER `deleted_at`;

ALTER TABLE `300sf`.`users` 
CHANGE COLUMN `timestamp` `timestamp` DATETIME NOT NULL ;

ALTER TABLE `300sf`.`api_key_tiers` 
CHANGE COLUMN `timestamp` `timestamp` DATETIME NOT NULL ;

ALTER TABLE `300sf`.`api_keys` 
CHANGE COLUMN `timestamp` `timestamp` DATETIME NOT NULL ;

CREATE TABLE `300sf`.`translations` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `en` VARCHAR(128) NULL,
  `sv` VARCHAR(128) NULL,
  `la` VARCHAR(128) NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE);

CREATE TABLE `300sf`.`genders` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_genders_1_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_genders_1`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`orders` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_orders_1_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_orders_1`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);
