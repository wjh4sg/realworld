CREATE TABLE IF NOT EXISTS user_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    bio TEXT,
    image VARCHAR(255),
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY idx_user_models_username (username),
    UNIQUE KEY idx_user_models_email (email),
    KEY idx_user_models_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS follow_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    following_id BIGINT UNSIGNED NOT NULL,
    followed_by_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_follow_models_following FOREIGN KEY (following_id) REFERENCES user_models(id) ON DELETE CASCADE,
    CONSTRAINT fk_follow_models_followed_by FOREIGN KEY (followed_by_id) REFERENCES user_models(id) ON DELETE CASCADE,
    UNIQUE KEY unique_follow (following_id, followed_by_id),
    KEY idx_follow_models_following_id (following_id),
    KEY idx_follow_models_followed_by_id (followed_by_id),
    KEY idx_follow_models_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS article_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    body TEXT,
    author_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_article_models_author FOREIGN KEY (author_id) REFERENCES user_models(id) ON DELETE CASCADE,
    UNIQUE KEY idx_article_models_slug (slug),
    KEY idx_article_models_author_id (author_id),
    KEY idx_article_models_deleted_at (deleted_at),
    KEY idx_article_created_at_desc (created_at DESC),
    KEY idx_article_created_at_id (created_at DESC, id),
    KEY idx_article_id_desc (id DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS tag_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tag VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY idx_tag_models_tag (tag),
    KEY idx_tag_models_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS article_tags (
    article_model_id BIGINT UNSIGNED NOT NULL,
    tag_model_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (article_model_id, tag_model_id),
    CONSTRAINT fk_article_tags_article FOREIGN KEY (article_model_id) REFERENCES article_models(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_tags_tag FOREIGN KEY (tag_model_id) REFERENCES tag_models(id) ON DELETE CASCADE,
    KEY idx_article_tags_tag_model_id (tag_model_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS favorite_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    favorite_id BIGINT UNSIGNED NOT NULL,
    favorite_by_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_favorite_models_article FOREIGN KEY (favorite_id) REFERENCES article_models(id) ON DELETE CASCADE,
    CONSTRAINT fk_favorite_models_user FOREIGN KEY (favorite_by_id) REFERENCES user_models(id) ON DELETE CASCADE,
    UNIQUE KEY unique_favorite (favorite_id, favorite_by_id),
    KEY idx_favorite_models_favorite_id (favorite_id),
    KEY idx_favorite_models_favorite_by_id (favorite_by_id),
    KEY idx_favorite_models_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS comment_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    article_id BIGINT UNSIGNED NOT NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_comment_models_article FOREIGN KEY (article_id) REFERENCES article_models(id) ON DELETE CASCADE,
    CONSTRAINT fk_comment_models_author FOREIGN KEY (author_id) REFERENCES user_models(id) ON DELETE CASCADE,
    KEY idx_comment_models_article_id (article_id),
    KEY idx_comment_models_author_id (author_id),
    KEY idx_comment_models_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
