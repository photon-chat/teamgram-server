-- City Activity tables for city-based event app
-- 2026-04-01

CREATE TABLE IF NOT EXISTS `activities` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `title` VARCHAR(255) NOT NULL DEFAULT '',
  `description` TEXT,
  `photo_id` BIGINT NOT NULL DEFAULT 0,
  `city` VARCHAR(100) NOT NULL DEFAULT '',
  `start_time` BIGINT NOT NULL DEFAULT 0,
  `end_time` BIGINT NOT NULL DEFAULT 0,
  `max_participants` INT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1=active 2=cancelled 3=finished',
  `is_global` TINYINT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_city_status` (`city`, `status`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_is_global` (`is_global`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `activity_participants` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `activity_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `city` VARCHAR(100) NOT NULL DEFAULT '',
  `joined_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_activity_user` (`activity_id`, `user_id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed default global activities
INSERT INTO `activities` (`user_id`, `title`, `description`, `city`, `start_time`, `end_time`, `status`, `is_global`, `created_at`, `updated_at`) VALUES
(0, '王者荣耀开黑', '找队友一起上分，不论段位，快乐游戏最重要！', '', 0, 0, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(0, '周末户外徒步', '一起去户外走走，呼吸新鲜空气，锻炼身体交朋友', '', 0, 0, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
