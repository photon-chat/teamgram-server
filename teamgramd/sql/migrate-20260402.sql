-- Activity media (photos) table
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `activity_media` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `activity_id` BIGINT NOT NULL,
  `photo_id` BIGINT NOT NULL,
  `sort_order` INT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  INDEX idx_activity_id (`activity_id`),
  UNIQUE INDEX idx_activity_photo (`activity_id`, `photo_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
