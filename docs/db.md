# データベース設計書 - 情報収集アプリ

## 1. 概要

### 1.1 システム概要
異なるアプリケーション（Twitter、Instagram、YouTube等）から日次で情報を取得し、Webで閲覧できるようにするアプリケーション。

### 1.2 設計方針
- **正規化**: 検索条件を完全に正規化し、型安全性を確保
- **拡張性**: 新しいデータソースの追加が容易
- **セキュリティ**: Row Level Security (RLS) によるユーザー間のデータ分離

## 2. テーブル設計

### 2.1 ユーザー管理

#### users
ユーザー情報を管理（Supabase Authと連携）

| カラム名 | 型 | 制約 | 説明 |
|---------|---|------|------|
| id | uuid | PRIMARY KEY, REFERENCES auth.users(id) | ユーザーID |
| email | text | UNIQUE, NOT NULL | メールアドレス |
| created_at | timestamp with time zone | DEFAULT now() | 作成日時 |

### 2.2 データソース管理

#### data_sources
対応するデータソース（Twitter、YouTube等）の定義

| カラム名 | 型 | 制約 | 説明 |
|---------|---|------|------|
| id | text | PRIMARY KEY | データソースID（'twitter', 'youtube'等） |
| name | text | NOT NULL | 表示名 |
| is_active | boolean | DEFAULT true | 有効/無効フラグ |
| created_at | timestamp with time zone | DEFAULT now() | 作成日時 |

### 2.3 検索条件定義

#### condition_types
検索条件タイプのマスタ定義

| カラム名 | 型 | 制約 | 説明 |
|---------|---|------|------|
| id | text | PRIMARY KEY | 条件タイプID（'account', 'keyword'等） |
| name | text | NOT NULL | 表示名（'アカウント', 'キーワード'等） |
| data_type | text | NOT NULL | データ型（'string', 'boolean', 'array'） |
| description | text | | 説明文 |

#### data_source_conditions
各データソースで使用可能な検索条件の関連

| カラム名 | 型 | 制約 | 説明 |
|---------|---|------|------|
| data_source_id | text | REFERENCES data_sources(id) | データソースID |
| condition_type_id | text | REFERENCES condition_types(id) | 条件タイプID |
| is_required | boolean | DEFAULT false | 必須フラグ |
| placeholder | text | | 入力欄のプレースホルダー |
| default_value | text | | デフォルト値 |
| PRIMARY KEY | | (data_source_id, condition_type_id) | 複合主キー |

### 2.4 ユーザー設定

#### user_fetch_configs
ユーザーの取得設定（基本情報）

| カラム名 | 型 | 制約 | 説明 |
|---------|---|------|------|
| id | uuid | PRIMARY KEY, DEFAULT gen_random_uuid() | 設定ID |
| user_id | uuid | NOT NULL, REFERENCES users(id) ON DELETE CASCADE | ユーザーID |
| data_source_id | text | NOT NULL, REFERENCES data_sources(id) | データソースID |
| name | text | | 設定名（ユーザーが付ける任意の名前） |
| is_active | boolean | DEFAULT true | 有効/無効フラグ |
| created_at | timestamp with time zone | DEFAULT now() | 作成日時 |
| updated_at | timestamp with time zone | DEFAULT now() | 更新日時 |

#### user_conditions
ユーザーの検索条件（詳細）

| カラム名 | 型 | 制約 | 説明 |
|---------|---|------|------|
| id | uuid | PRIMARY KEY, DEFAULT gen_random_uuid() | 条件ID |
| config_id | uuid | REFERENCES user_fetch_configs(id) ON DELETE CASCADE | 設定ID |
| condition_type_id | text | REFERENCES condition_types(id) | 条件タイプID |
| value | text | | 値（単一値の場合） |
| created_at | timestamp with time zone | DEFAULT now() | 作成日時 |

#### user_condition_items
配列型の検索条件の値（複数アカウント、複数キーワード等）

| カラム名 | 型 | 制約 | 説明 |
|---------|---|------|------|
| id | uuid | PRIMARY KEY, DEFAULT gen_random_uuid() | アイテムID |
| condition_id | uuid | REFERENCES user_conditions(id) ON DELETE CASCADE | 条件ID |
| value | text | NOT NULL | 値 |
| sort_order | integer | DEFAULT 0 | 並び順 |

### 2.5 取得データ

#### fetched_data
取得したデータ（構造化済み）

| カラム名 | 型 | 制約 | 説明 |
|---------|---|------|------|
| id | uuid | PRIMARY KEY, DEFAULT gen_random_uuid() | データID |
| config_id | uuid | NOT NULL, REFERENCES user_fetch_configs(id) ON DELETE CASCADE | 設定ID |
| title | text | NOT NULL | タイトル |
| content | text | | 本文 |
| url | text | | URL |
| author_name | text | | 投稿者名 |
| author_id | text | | 投稿者ID |
| author_avatar_url | text | | 投稿者アバターURL |
| published_at | timestamp with time zone | | 公開日時 |
| engagement | jsonb | DEFAULT '{}' | エンゲージメント（likes, comments等） |
| media | jsonb | DEFAULT '[]' | メディアURL配列 |
| tags | text[] | DEFAULT '{}' | タグ配列 |
| raw_data | jsonb | | 元データ（デバッグ用） |
| fetched_at | timestamp with time zone | DEFAULT now() | 取得日時 |
| created_at | timestamp with time zone | DEFAULT now() | 作成日時 |

## 3. インデックス

```sql
-- user_fetch_configs
CREATE INDEX idx_user_fetch_configs_user_id ON user_fetch_configs(user_id);
CREATE INDEX idx_user_fetch_configs_data_source_id ON user_fetch_configs(data_source_id);

-- user_conditions
CREATE INDEX idx_user_conditions_config_id ON user_conditions(config_id);

-- user_condition_items
CREATE INDEX idx_user_condition_items_condition_id ON user_condition_items(condition_id);

-- fetched_data
CREATE INDEX idx_fetched_data_config_id ON fetched_data(config_id);
CREATE INDEX idx_fetched_data_fetched_at ON fetched_data(fetched_at DESC);
CREATE INDEX idx_fetched_data_published_at ON fetched_data(published_at DESC);
```

## 4. ビュー

### fetched_data_view
UIで使いやすいように結合済みのビュー

```sql
CREATE VIEW fetched_data_view AS
SELECT 
  fd.*,
  ufc.user_id,
  ufc.data_source_id,
  ds.name as data_source_name,
  u.email as user_email
FROM fetched_data fd
JOIN user_fetch_configs ufc ON fd.config_id = ufc.id
JOIN data_sources ds ON ufc.data_source_id = ds.id
JOIN users u ON ufc.user_id = u.id;
```

## 5. Row Level Security (RLS)

### user_fetch_configs
```sql
ALTER TABLE user_fetch_configs ENABLE ROW LEVEL SECURITY;

-- SELECT: 自分の設定のみ参照可能
CREATE POLICY "Users can view own configs" ON user_fetch_configs
  FOR SELECT USING (auth.uid() = user_id);

-- INSERT: 自分の設定のみ作成可能
CREATE POLICY "Users can insert own configs" ON user_fetch_configs
  FOR INSERT WITH CHECK (auth.uid() = user_id);

-- UPDATE: 自分の設定のみ更新可能
CREATE POLICY "Users can update own configs" ON user_fetch_configs
  FOR UPDATE USING (auth.uid() = user_id);

-- DELETE: 自分の設定のみ削除可能
CREATE POLICY "Users can delete own configs" ON user_fetch_configs
  FOR DELETE USING (auth.uid() = user_id);
```

### user_conditions
```sql
ALTER TABLE user_conditions ENABLE ROW LEVEL SECURITY;

-- 設定の所有者のみアクセス可能
CREATE POLICY "Users can manage own conditions" ON user_conditions
  FOR ALL USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = user_conditions.config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );
```

### fetched_data
```sql
ALTER TABLE fetched_data ENABLE ROW LEVEL SECURITY;

-- 自分の設定で取得したデータのみ参照可能
CREATE POLICY "Users can view own data" ON fetched_data
  FOR SELECT USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = fetched_data.config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );
```

## 6. 初期データ

### データソース
```sql
INSERT INTO data_sources (id, name) VALUES
  ('twitter', 'Twitter'),
  ('instagram', 'Instagram'),
  ('youtube', 'YouTube'),
  ('hackernews', 'Hacker News'),
  ('reddit', 'Reddit');
```

### 条件タイプ
```sql
INSERT INTO condition_types (id, name, data_type, description) VALUES
  ('account', 'アカウント', 'array', 'フォローするアカウント'),
  ('keyword', 'キーワード', 'array', '検索キーワード'),
  ('channel', 'チャンネル', 'array', '購読チャンネル'),
  ('subreddit', 'Subreddit', 'array', '購読するSubreddit'),
  ('exclude_retweets', 'リツイートを除外', 'boolean', 'リツイートを検索結果から除外'),
  ('min_score', '最小スコア', 'string', '最小スコア（数値）');
```

### データソースと条件の関連
```sql
-- Twitter
INSERT INTO data_source_conditions (data_source_id, condition_type_id, placeholder) VALUES
  ('twitter', 'account', '@username'),
  ('twitter', 'keyword', '検索キーワード'),
  ('twitter', 'exclude_retweets', null);

-- YouTube
INSERT INTO data_source_conditions (data_source_id, condition_type_id, placeholder) VALUES
  ('youtube', 'channel', 'チャンネルID or URL'),
  ('youtube', 'keyword', 'キーワード');

-- Instagram
INSERT INTO data_source_conditions (data_source_id, condition_type_id, placeholder) VALUES
  ('instagram', 'account', 'username（@なし）');

-- Reddit
INSERT INTO data_source_conditions (data_source_id, condition_type_id, placeholder) VALUES
  ('reddit', 'subreddit', 'subreddit名'),
  ('reddit', 'keyword', 'キーワード'),
  ('reddit', 'min_score', '0');

-- Hacker News
INSERT INTO data_source_conditions (data_source_id, condition_type_id, placeholder) VALUES
  ('hackernews', 'keyword', 'キーワード'),
  ('hackernews', 'min_score', '10');
```

## 7. 使用例

### ユーザーがTwitterの検索条件を設定する場合

1. user_fetch_configsに基本設定を作成
```sql
INSERT INTO user_fetch_configs (user_id, data_source_id, name) 
VALUES ('user-uuid', 'twitter', '技術系アカウント') 
RETURNING id;  -- 例: 'config-uuid'
```

2. user_conditionsに条件を作成
```sql
-- アカウント条件
INSERT INTO user_conditions (config_id, condition_type_id) 
VALUES ('config-uuid', 'account') 
RETURNING id;  -- 例: 'condition1-uuid'

-- キーワード条件
INSERT INTO user_conditions (config_id, condition_type_id) 
VALUES ('config-uuid', 'keyword') 
RETURNING id;  -- 例: 'condition2-uuid'

-- リツイート除外
INSERT INTO user_conditions (config_id, condition_type_id, value) 
VALUES ('config-uuid', 'exclude_retweets', 'true');
```

3. user_condition_itemsに具体的な値を設定
```sql
-- アカウント
INSERT INTO user_condition_items (condition_id, value, sort_order) VALUES
  ('condition1-uuid', '@golang', 0),
  ('condition1-uuid', '@rustlang', 1);

-- キーワード
INSERT INTO user_condition_items (condition_id, value, sort_order) VALUES
  ('condition2-uuid', 'typescript', 0),
  ('condition2-uuid', 'react', 1);
```

### バッチ処理での検索条件取得

```sql
-- 全ユーザーの有効な設定を取得
SELECT 
  ufc.id as config_id,
  ufc.user_id,
  ufc.data_source_id,
  uc.condition_type_id,
  uc.value,
  array_agg(uci.value ORDER BY uci.sort_order) as array_values
FROM user_fetch_configs ufc
LEFT JOIN user_conditions uc ON ufc.id = uc.config_id
LEFT JOIN user_condition_items uci ON uc.id = uci.condition_id
WHERE ufc.is_active = true
GROUP BY ufc.id, ufc.user_id, ufc.data_source_id, uc.condition_type_id, uc.value;
```

## 8. 今後の拡張性

- 新しいデータソースの追加
  1. data_sourcesテーブルに1行追加
  2. condition_typesに必要な条件タイプを追加（既存のものを再利用可能）
  3. data_source_conditionsで関連付け

- 新しい検索条件の追加
  1. condition_typesに条件タイプを追加
  2. data_source_conditionsで必要なデータソースと関連付け

この設計により、JSONを使わずに完全に正規化された形で検索条件を管理できます。
