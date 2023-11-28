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
  `revoked_at` DATETIME NULL,
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

CREATE TABLE `300sf`.`activities` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_activities_1_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_activities_1`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`methods` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_methods_1_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_methods_1`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`ages` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_ages_1_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_ages_1`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`species` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  `order_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_species_1_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_species_1`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  INDEX `fk_species_2_idx` (`order_id` ASC) VISIBLE,
  CONSTRAINT `fk_species_2`
    FOREIGN KEY (`order_id`)
    REFERENCES `300sf`.`orders` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`list_templates` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  `user_id` INT NOT NULL,
  `visibility` INT NOT NULL,
  `archived` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_list_templates_1_idx` (`user_id` ASC) VISIBLE,
  INDEX `fk_list_templates_2_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_list_templates_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_list_templates_2`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`list_template_species` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `list_template_id` INT NOT NULL,
  `species_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_list_template_species_1_idx` (`list_template_id` ASC) VISIBLE,
  INDEX `fk_list_template_species_2_idx` (`species_id` ASC) VISIBLE,
  UNIQUE INDEX `index4` (`list_template_id` ASC, `species_id` ASC) VISIBLE,
  CONSTRAINT `fk_list_template_species_1`
    FOREIGN KEY (`list_template_id`)
    REFERENCES `300sf`.`list_templates` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_list_template_species_2`
    FOREIGN KEY (`species_id`)
    REFERENCES `300sf`.`species` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`lists` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  `user_id` INT NOT NULL,
  `visibility` INT NOT NULL,
  `archived` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_lists_1_idx` (`user_id` ASC) VISIBLE,
  INDEX `fk_lists_2_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_lists_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_lists_2`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

ALTER TABLE `300sf`.`lists` 
  ADD COLUMN `from` DATETIME NULL AFTER `archived`,
  ADD COLUMN `until` DATETIME NULL AFTER `from`;

CREATE TABLE `300sf`.`list_species` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `list_id` INT NOT NULL,
  `species_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `index2` (`list_id` ASC, `species_id` ASC) VISIBLE,
  INDEX `fk_list_species_2_idx` (`species_id` ASC) VISIBLE,
  CONSTRAINT `fk_list_species_1`
    FOREIGN KEY (`list_id`)
    REFERENCES `300sf`.`lists` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_list_species_2`
    FOREIGN KEY (`species_id`)
    REFERENCES `300sf`.`species` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`reports` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `user_id` INT NOT NULL,
  `species_id` INT NOT NULL,
  `from` DATETIME NULL,
  `until` DATETIME NULL,
  `latitude` FLOAT NULL,
  `longitude` FLOAT NULL,
  `quantity` INT NULL,
  `age_id` INT NULL,
  `activity_id` INT NULL,
  `gender_id` INT NULL,
  `method_id` INT NULL,
  `comment` TEXT NULL,
  `archived` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_reports_1_idx` (`activity_id` ASC) VISIBLE,
  INDEX `fk_reports_2_idx` (`user_id` ASC) VISIBLE,
  INDEX `fk_reports_3_idx` (`species_id` ASC) VISIBLE,
  INDEX `fk_reports_4_idx` (`gender_id` ASC) VISIBLE,
  INDEX `fk_reports_5_idx` (`method_id` ASC) VISIBLE,
  CONSTRAINT `fk_reports_1`
    FOREIGN KEY (`activity_id`)
    REFERENCES `300sf`.`activities` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_reports_2`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_reports_3`
    FOREIGN KEY (`species_id`)
    REFERENCES `300sf`.`species` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_reports_4`
    FOREIGN KEY (`gender_id`)
    REFERENCES `300sf`.`genders` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_reports_5`
    FOREIGN KEY (`age_id`)
    REFERENCES `300sf`.`ages` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_reports_6`
    FOREIGN KEY (`method_id`)
    REFERENCES `300sf`.`methods` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`list_reports` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `list_id` INT NOT NULL,
  `report_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `index2` (`list_id` ASC, `report_id` ASC) VISIBLE,
  INDEX `fk_list_reports_2_idx` (`report_id` ASC) VISIBLE,
  CONSTRAINT `fk_list_reports_1`
    FOREIGN KEY (`list_id`)
    REFERENCES `300sf`.`lists` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_list_reports_2`
    FOREIGN KEY (`report_id`)
    REFERENCES `300sf`.`reports` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

ALTER TABLE `300sf`.`list_templates` 
ADD COLUMN `description_id` INT NULL AFTER `archived`,
ADD INDEX `fk_list_templates_3_idx` (`description_id` ASC) VISIBLE;
;
ALTER TABLE `300sf`.`list_templates` 
ADD CONSTRAINT `fk_list_templates_3`
  FOREIGN KEY (`description_id`)
  REFERENCES `300sf`.`translations` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION;

CREATE TABLE `300sf`.`tags` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `user_id` INT NOT NULL,
  `color` VARCHAR(9) NULL,
  `archived` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_tags_1_idx` (`user_id` ASC) VISIBLE,
  CONSTRAINT `fk_tags_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`report_tags` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `report_id` INT NOT NULL,
  `tag_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `index2` (`report_id` ASC, `tag_id` ASC) VISIBLE,
  INDEX `fk_report_tags_2_idx` (`tag_id` ASC) VISIBLE,
  CONSTRAINT `fk_report_tags_1`
    FOREIGN KEY (`report_id`)
    REFERENCES `300sf`.`reports` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_report_tags_2`
    FOREIGN KEY (`tag_id`)
    REFERENCES `300sf`.`tags` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

ALTER TABLE `300sf`.`lists` 
ADD COLUMN `list_template_id` INT NULL AFTER `until`,
ADD INDEX `fk_lists_3_idx` (`list_template_id` ASC) VISIBLE;
;
ALTER TABLE `300sf`.`lists` 
ADD CONSTRAINT `fk_lists_3`
  FOREIGN KEY (`list_template_id`)
  REFERENCES `300sf`.`list_templates` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION;

CREATE TABLE `300sf`.`comments` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `user_id` INT NOT NULL,
  `text` TEXT NOT NULL,
  `edited` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_comment_1_idx` (`user_id` ASC) VISIBLE,
  CONSTRAINT `fk_comment_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

ALTER TABLE `300sf`.`reports` 
CHANGE COLUMN `comment` `comment_id` INT NULL DEFAULT NULL ;

ALTER TABLE `300sf`.`reports` 
ADD INDEX `fk_reports_7_idx` (`comment_id` ASC) VISIBLE;
;
ALTER TABLE `300sf`.`reports` 
ADD CONSTRAINT `fk_reports_7`
  FOREIGN KEY (`comment_id`)
  REFERENCES `300sf`.`comments` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION;

CREATE TABLE `300sf`.`notifications` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `report_id` INT NOT NULL,
  `comment_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  UNIQUE INDEX `index4` (`report_id` ASC, `comment_id` ASC) VISIBLE,
  INDEX `fk_notifications_2_idx` (`comment_id` ASC) VISIBLE,
  CONSTRAINT `fk_notifications_1`
    FOREIGN KEY (`report_id`)
    REFERENCES `300sf`.`reports` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_notifications_2`
    FOREIGN KEY (`comment_id`)
    REFERENCES `300sf`.`comments` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`notification_comments` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `notification_id` INT NOT NULL,
  `comment_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `index2` (`notification_id` ASC, `comment_id` ASC) VISIBLE,
  INDEX `fk_notification_comments_2_idx` (`comment_id` ASC) VISIBLE,
  CONSTRAINT `fk_notification_comments_1`
    FOREIGN KEY (`notification_id`)
    REFERENCES `300sf`.`notifications` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_notification_comments_2`
    FOREIGN KEY (`comment_id`)
    REFERENCES `300sf`.`comments` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

ALTER TABLE `300sf`.`notifications` 
ADD COLUMN `processed` DATETIME NULL AFTER `comment_id`;

CREATE TABLE `300sf`.`zones` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name_id` INT NOT NULL,
  `polygon` TEXT NOT NULL,
  `admin_level` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_zones_1_idx` (`name_id` ASC) VISIBLE,
  CONSTRAINT `fk_zones_1`
    FOREIGN KEY (`name_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`subscriptions` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(45) NOT NULL,
  `user_id` INT NOT NULL,
  `zone_id` INT NOT NULL,
  `paused` DATETIME NULL,
  `archived` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_subscriptions_1_idx` (`user_id` ASC) VISIBLE,
  INDEX `fk_subscriptions_2_idx` (`zone_id` ASC) VISIBLE,
  CONSTRAINT `fk_subscriptions_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_subscriptions_2`
    FOREIGN KEY (`zone_id`)
    REFERENCES `300sf`.`zones` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`subscription_species` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `subscription_id` INT NOT NULL,
  `species_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `index2` (`subscription_id` ASC, `species_id` ASC) VISIBLE,
  INDEX `fk_subscription_species_2_idx` (`species_id` ASC) VISIBLE,
  CONSTRAINT `fk_subscription_species_1`
    FOREIGN KEY (`subscription_id`)
    REFERENCES `300sf`.`subscriptions` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_subscription_species_2`
    FOREIGN KEY (`species_id`)
    REFERENCES `300sf`.`species` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

ALTER TABLE `300sf`.`zones` 
ADD COLUMN `parent_zone_id` INT NULL AFTER `admin_level`,
ADD INDEX `fk_zones_2_idx` (`parent_zone_id` ASC) VISIBLE;
;
ALTER TABLE `300sf`.`zones` 
ADD CONSTRAINT `fk_zones_2`
  FOREIGN KEY (`parent_zone_id`)
  REFERENCES `300sf`.`zones` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION;

  ALTER TABLE `300sf`.`zones` 
CHANGE COLUMN `polygon` `polygon` MEDIUMBLOB NOT NULL ;

ALTER TABLE `300sf`.`reports` 
ADD COLUMN `zone_id` INT NULL AFTER `longitude`,
ADD INDEX `fk_reports_8_idx` (`zone_id` ASC) VISIBLE;
;
ALTER TABLE `300sf`.`reports` 
ADD CONSTRAINT `fk_reports_8`
  FOREIGN KEY (`zone_id`)
  REFERENCES `300sf`.`zones` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION;

  
CREATE TABLE `300sf`.`rarities` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `species_id` INT NOT NULL,
  `zone_id` INT NOT NULL,
  `category` CHAR NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `index2` (`species_id` ASC, `zone_id` ASC) VISIBLE,
  INDEX `fk_rarities_2_idx` (`species_id` ASC) VISIBLE,
  CONSTRAINT `fk_rarities_1`
    FOREIGN KEY (`species_id`)
    REFERENCES `300sf`.`species` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_rarities_2`
    FOREIGN KEY (`zone_id`)
    REFERENCES `300sf`.`zones` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`emails` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `from_email` VARCHAR(254) NOT NULL,
  `from_name` VARCHAR(128) NOT NULL,
  `to_email` VARCHAR(254) NOT NULL,
  `to_name` VARCHAR(128) NOT NULL,
  `subject` VARCHAR(128) NOT NULL,
  `text_part` TEXT NOT NULL,
  `html_part` TEXT NOT NULL,
  `send_at` DATETIME NOT NULL,
  `sent_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE);

CREATE TABLE `300sf`.`email_templates` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `from_email` VARCHAR(254) NOT NULL,
  `from_name` VARCHAR(128) NOT NULL,
  `subject_id` INT NOT NULL,
  `text_part_id` INT NOT NULL,
  `html_part_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_email_templates_1_idx` (`subject_id` ASC) VISIBLE,
  INDEX `fk_email_templates_2_idx` (`text_part_id` ASC) VISIBLE,
  INDEX `fk_email_templates_3_idx` (`html_part_id` ASC) VISIBLE,
  CONSTRAINT `fk_email_templates_1`
    FOREIGN KEY (`subject_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_email_templates_2`
    FOREIGN KEY (`text_part_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_email_templates_3`
    FOREIGN KEY (`html_part_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

ALTER TABLE `300sf`.`lists` 
ADD COLUMN `color` VARCHAR(16) NULL AFTER `list_template_id`;

ALTER TABLE `300sf`.`users` 
ADD COLUMN `imported_at` DATETIME NULL AFTER `accepted_terms_at`;

CREATE TABLE `300sf`.`language` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `iso_name` VARCHAR(256) NOT NULL,
  `native_name` VARCHAR(256) NOT NULL,
  `code` VARCHAR(16) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `code_UNIQUE` (`code` ASC) VISIBLE);

ALTER TABLE `300sf`.`users` 
ADD COLUMN `language_id` INT AFTER `imported_at`,
ADD INDEX `fk_users_2_idx` (`language_id` ASC) VISIBLE;
;
ALTER TABLE `300sf`.`users` 
ADD CONSTRAINT `fk_users_2`
  FOREIGN KEY (`language_id`)
  REFERENCES `300sf`.`language` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`language` 
RENAME TO  `300sf`.`languages` ;

INSERT INTO `300sf`.`languages` (`iso_name`, `native_name`, `code`) VALUES ('English', 'English', 'en');
INSERT INTO `300sf`.`languages` (`iso_name`, `native_name`, `code`) VALUES ('Swedish', 'Svenska', 'sv');

ALTER TABLE `300sf`.`users` 
ADD COLUMN `portrait` VARCHAR(256) NULL AFTER `language_id`;

ALTER TABLE `300sf`.`users` 
DROP FOREIGN KEY `fk_users_1`;
ALTER TABLE `300sf`.`users` 
DROP COLUMN `icon_id`,
DROP INDEX `fk_users_1_idx` ;
;

INSERT INTO `300sf`.`languages` (`iso_name`, `native_name`, `code`) VALUES ('German', 'Deutch', 'de');

ALTER TABLE `300sf`.`translations` 
ADD COLUMN `de` VARCHAR(128) NULL AFTER `la`;

ALTER TABLE `300sf`.`emails` 
ADD COLUMN `checksum` VARCHAR(64) NOT NULL AFTER `sent_at`;



CREATE TABLE `300sf`.`change_password_requests` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `expire_at` DATETIME NOT NULL,
  `token` VARCHAR(256) NOT NULL,
  `user_id` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  UNIQUE INDEX `token_UNIQUE` (`token` ASC) VISIBLE,
  INDEX `fk_change_password_requests_1_idx` (`user_id` ASC) VISIBLE,
  CONSTRAINT `fk_change_password_requests_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

CREATE TABLE `300sf`.`delete_account_requests` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `expire_at` DATETIME NOT NULL,
  `token` VARCHAR(256) NOT NULL,
  `user_id` INT NOT NULL,
  `delete_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  UNIQUE INDEX `token_UNIQUE` (`token` ASC) VISIBLE,
  INDEX `fk_delete_account_requests_1_idx` (`user_id` ASC) VISIBLE,
  CONSTRAINT `fk_delete_account_requests_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

    ALTER TABLE `300sf`.`delete_account_requests` 
CHANGE COLUMN `delete_at` `delete_at` DATETIME NULL ;

ALTER TABLE `300sf`.`users` 
ADD COLUMN `delete_account_request_id` INT NULL AFTER `portrait`,
ADD INDEX `fk_users_1_idx` (`delete_account_request_id` ASC) VISIBLE;
;
ALTER TABLE `300sf`.`users` 
ADD CONSTRAINT `fk_users_1`
  FOREIGN KEY (`delete_account_request_id`)
  REFERENCES `300sf`.`delete_account_requests` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`api_keys` 
DROP FOREIGN KEY `fk_api_keys_2`;
ALTER TABLE `300sf`.`api_keys` 
ADD CONSTRAINT `fk_api_keys_2`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;


ALTER TABLE `300sf`.`change_password_requests` 
DROP FOREIGN KEY `fk_change_password_requests_1`;
ALTER TABLE `300sf`.`change_password_requests` 
ADD CONSTRAINT `fk_change_password_requests_1`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`comments` 
DROP FOREIGN KEY `fk_comment_1`;
ALTER TABLE `300sf`.`comments` 
ADD CONSTRAINT `fk_comment_1`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`reports` 
DROP FOREIGN KEY `fk_reports_2`;
ALTER TABLE `300sf`.`reports` 
ADD CONSTRAINT `fk_reports_2`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`reports` 
DROP FOREIGN KEY `fk_reports_2`,
DROP FOREIGN KEY `fk_reports_7`;
ALTER TABLE `300sf`.`reports` 
ADD CONSTRAINT `fk_reports_2`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION,
ADD CONSTRAINT `fk_reports_7`
  FOREIGN KEY (`comment_id`)
  REFERENCES `300sf`.`comments` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`list_reports` 
DROP FOREIGN KEY `fk_list_reports_2`;
ALTER TABLE `300sf`.`list_reports` 
ADD CONSTRAINT `fk_list_reports_2`
  FOREIGN KEY (`report_id`)
  REFERENCES `300sf`.`reports` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`notifications` 
DROP FOREIGN KEY `fk_notifications_1`;
ALTER TABLE `300sf`.`notifications` 
ADD CONSTRAINT `fk_notifications_1`
  FOREIGN KEY (`report_id`)
  REFERENCES `300sf`.`reports` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`report_tags` 
DROP FOREIGN KEY `fk_report_tags_1`;
ALTER TABLE `300sf`.`report_tags` 
ADD CONSTRAINT `fk_report_tags_1`
  FOREIGN KEY (`report_id`)
  REFERENCES `300sf`.`reports` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`delete_account_requests` 
DROP FOREIGN KEY `fk_delete_account_requests_1`;
ALTER TABLE `300sf`.`delete_account_requests` 
ADD CONSTRAINT `fk_delete_account_requests_1`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`users` 
DROP FOREIGN KEY `fk_users_1`;
ALTER TABLE `300sf`.`users` 
ADD CONSTRAINT `fk_users_1`
  FOREIGN KEY (`delete_account_request_id`)
  REFERENCES `300sf`.`delete_account_requests` (`id`)
  ON DELETE SET NULL
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`list_templates` 
DROP FOREIGN KEY `fk_list_templates_1`;
ALTER TABLE `300sf`.`list_templates` 
ADD CONSTRAINT `fk_list_templates_1`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`list_template_species` 
DROP FOREIGN KEY `fk_list_template_species_1`;
ALTER TABLE `300sf`.`list_template_species` 
ADD CONSTRAINT `fk_list_template_species_1`
  FOREIGN KEY (`list_template_id`)
  REFERENCES `300sf`.`list_templates` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`lists` 
DROP FOREIGN KEY `fk_lists_3`;
ALTER TABLE `300sf`.`lists` 
ADD CONSTRAINT `fk_lists_3`
  FOREIGN KEY (`list_template_id`)
  REFERENCES `300sf`.`list_templates` (`id`)
  ON DELETE SET NULL
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`lists` 
DROP FOREIGN KEY `fk_lists_1`;
ALTER TABLE `300sf`.`lists` 
ADD CONSTRAINT `fk_lists_1`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`list_species` 
DROP FOREIGN KEY `fk_list_species_1`;
ALTER TABLE `300sf`.`list_species` 
ADD CONSTRAINT `fk_list_species_1`
  FOREIGN KEY (`list_id`)
  REFERENCES `300sf`.`lists` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`tags` 
DROP FOREIGN KEY `fk_tags_1`;
ALTER TABLE `300sf`.`tags` 
ADD CONSTRAINT `fk_tags_1`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`list_reports` 
DROP FOREIGN KEY `fk_list_reports_1`;
ALTER TABLE `300sf`.`list_reports` 
ADD CONSTRAINT `fk_list_reports_1`
  FOREIGN KEY (`list_id`)
  REFERENCES `300sf`.`lists` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`notification_comments` 
DROP FOREIGN KEY `fk_notification_comments_2`;
ALTER TABLE `300sf`.`notification_comments` 
ADD CONSTRAINT `fk_notification_comments_2`
  FOREIGN KEY (`comment_id`)
  REFERENCES `300sf`.`comments` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`notifications` 
DROP FOREIGN KEY `fk_notifications_2`;
ALTER TABLE `300sf`.`notifications` 
ADD CONSTRAINT `fk_notifications_2`
  FOREIGN KEY (`comment_id`)
  REFERENCES `300sf`.`comments` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`report_tags` 
DROP FOREIGN KEY `fk_report_tags_2`;
ALTER TABLE `300sf`.`report_tags` 
ADD CONSTRAINT `fk_report_tags_2`
  FOREIGN KEY (`tag_id`)
  REFERENCES `300sf`.`tags` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`subscriptions` 
DROP FOREIGN KEY `fk_subscriptions_1`;
ALTER TABLE `300sf`.`subscriptions` 
ADD CONSTRAINT `fk_subscriptions_1`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`subscription_species` 
DROP FOREIGN KEY `fk_subscription_species_1`;
ALTER TABLE `300sf`.`subscription_species` 
ADD CONSTRAINT `fk_subscription_species_1`
  FOREIGN KEY (`subscription_id`)
  REFERENCES `300sf`.`subscriptions` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;


ALTER TABLE `300sf`.`emails` 
ADD COLUMN `reply_email` VARCHAR(254) NOT NULL AFTER `checksum`,
ADD COLUMN `reply_name` VARCHAR(128) NOT NULL AFTER `reply_email`;


CREATE TABLE `300sf`.`announcements` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `timestamp` DATETIME NOT NULL,
  `uuid` VARCHAR(32) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `content_id` INT NOT NULL,
  `from` DATETIME NULL,
  `until` DATETIME NULL,
  `priority` INT NOT NULL,
  `user_id` INT NULL,
  `on_hold` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uuid_UNIQUE` (`uuid` ASC) VISIBLE,
  INDEX `index3` (`uuid` ASC) VISIBLE,
  INDEX `fk_announcements_1_idx` (`user_id` ASC) VISIBLE,
  INDEX `fk_announcements_2_idx` (`content_id` ASC) VISIBLE,
  CONSTRAINT `fk_announcements_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `300sf`.`users` (`id`)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_announcements_2`
    FOREIGN KEY (`content_id`)
    REFERENCES `300sf`.`translations` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);


CREATE TABLE `300sf`.`announcement_users` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `announcement_id` INT NOT NULL,
  `user_id` INT NOT NULL,
  PRIMARY KEY (`id`));

ALTER TABLE `300sf`.`announcement_users` 
ADD INDEX `fk_announcement_users_1_idx` (`announcement_id` ASC) VISIBLE,
ADD INDEX `fk_announcement_users_2_idx` (`user_id` ASC) VISIBLE;
;
ALTER TABLE `300sf`.`announcement_users` 
ADD CONSTRAINT `fk_announcement_users_1`
  FOREIGN KEY (`announcement_id`)
  REFERENCES `300sf`.`announcements` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION,
ADD CONSTRAINT `fk_announcement_users_2`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE NO ACTION
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`announcement_users` 
DROP FOREIGN KEY `fk_announcement_users_1`;
ALTER TABLE `300sf`.`announcement_users` 
ADD CONSTRAINT `fk_announcement_users_1`
  FOREIGN KEY (`announcement_id`)
  REFERENCES `300sf`.`announcements` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`announcement_users` 
DROP FOREIGN KEY `fk_announcement_users_2`;
ALTER TABLE `300sf`.`announcement_users` 
ADD CONSTRAINT `fk_announcement_users_2`
  FOREIGN KEY (`user_id`)
  REFERENCES `300sf`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE NO ACTION;

ALTER TABLE `300sf`.`translations` 
CHANGE COLUMN `en` `en` VARCHAR(512) NULL DEFAULT NULL ,
CHANGE COLUMN `sv` `sv` VARCHAR(512) NULL DEFAULT NULL ,
CHANGE COLUMN `la` `la` VARCHAR(512) NULL DEFAULT NULL ,
CHANGE COLUMN `de` `de` VARCHAR(512) NULL DEFAULT NULL ;

ALTER TABLE `300sf`.`announcements` 
DROP COLUMN `priority`;

ALTER TABLE `300sf`.`announcements` 
ADD COLUMN `color` VARCHAR(16) NOT NULL AFTER `on_hold`,
ADD COLUMN `icon` VARCHAR(32) NOT NULL AFTER `color`;

ALTER TABLE `300sf`.`api_keys` 
ADD COLUMN `client_id` VARCHAR(32) NOT NULL AFTER `name`;
ALTER TABLE `300sf`.`api_keys` 
ADD UNIQUE INDEX `client_id_UNIQUE` (`client_id` ASC) VISIBLE;
;

ALTER TABLE `300sf`.`users` 
ADD COLUMN `scopes` TEXT NOT NULL AFTER `delete_account_request_id`;

ALTER TABLE `300sf`.`api_keys` 
DROP COLUMN `scopes`;

ALTER TABLE `300sf`.`api_keys` 
ADD COLUMN `scopes` TEXT NOT NULL AFTER `revoked_at`;

ALTER TABLE `300sf`.`users` 
ADD COLUMN `username` VARCHAR(64) NOT NULL AFTER `email`;

ALTER TABLE `300sf`.`users` 
ADD UNIQUE INDEX `username_UNIQUE` (`username` ASC) VISIBLE;
;

ALTER TABLE `300sf`.`users` 
DROP COLUMN `permissions`;

-->