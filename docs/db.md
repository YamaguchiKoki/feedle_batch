# データベース設計書 - 情報収集アプリ

## 1. 概要

### 1.1 システム概要
異なるアプリケーション（Reddit、Twitter、YouTube等）から日次で情報を取得し、Webで閲覧できるようにするアプリケーション。

### 1.2 設計方針
- **シンプルな構造**: 必要最小限のテーブル構成
- **拡張性**: 新しいデータソースの追加が容易
- **セキュリティ**: Row Level Security (RLS) によるユーザー間のデータ分離

## 2. テーブル設計

### 2.1 ユーザー管理

#### users
ユーザー情報を管理（Supabase Authと連携）

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY, -- auth.users.id を参照
    name TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL
);
```

### 2.2 データソース定義

#### data_sources
対応するデータソース（Reddit、Twitter等）の定義

```sql
CREATE TABLE data_sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    icon TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 2.3 ユーザーの取得設定

#### user_fetch_configs
ユーザーごとの共通取得設定（親テーブル）

```sql
CREATE TABLE user_fetch_configs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    data_source_id TEXT NOT NULL REFERENCES data_sources(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- インデックス
CREATE INDEX user_fetch_configs_user_id_idx ON user_fetch_configs(user_id);
CREATE INDEX user_fetch_configs_data_source_idx ON user_fetch_configs(data_source_id);
```

#### reddit_fetch_configs
Reddit固有の取得設定

```sql
CREATE TABLE reddit_fetch_configs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_fetch_config_id UUID NOT NULL REFERENCES user_fetch_configs(id) ON DELETE CASCADE,
    subreddit TEXT,
    sort_by TEXT DEFAULT 'hot', -- hot, new, top, rising
    time_filter TEXT DEFAULT 'day', -- hour, day, week, month, year, all
    limit_count INTEGER DEFAULT 25,
    keywords TEXT[],
    created_at TIMESTAMP DEFAULT NOW()
);

-- インデックス
CREATE INDEX reddit_fetch_configs_user_fetch_config_id_idx ON reddit_fetch_configs(user_fetch_config_id);
```

#### youtube_fetch_configs
YouTube固有の取得設定

```sql
CREATE TABLE youtube_fetch_configs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_fetch_config_id UUID NOT NULL REFERENCES user_fetch_configs(id) ON DELETE CASCADE,
    channel_id TEXT,
    playlist_id TEXT,
    keywords TEXT[],
    max_results INTEGER DEFAULT 50,
    order_by TEXT DEFAULT 'relevance', -- relevance, date, viewCount, rating, title
    published_after TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- インデックス
CREATE INDEX youtube_fetch_configs_user_fetch_config_id_idx ON youtube_fetch_configs(user_fetch_config_id);
```

### 2.4 取得したデータ

#### fetched_data
各データソースから取得したデータ

```sql
CREATE TABLE fetched_data (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    config_id UUID NOT NULL REFERENCES user_fetch_configs(id),
    source TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    url TEXT,
    author_name TEXT,
    source_item_id TEXT,
    published_at TIMESTAMP,
    tags TEXT[] DEFAULT '{}',
    media_urls TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    fetched_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- 同一設定・ソース・アイテムIDの組み合わせでユニーク制約
    CONSTRAINT unique_source_item_per_config UNIQUE (config_id, source, source_item_id)
);

-- インデックス
CREATE INDEX fetched_data_config_id_idx ON fetched_data(config_id);
CREATE INDEX fetched_data_source_idx ON fetched_data(source);
CREATE INDEX fetched_data_published_at_idx ON fetched_data(published_at);
CREATE INDEX fetched_data_fetched_at_idx ON fetched_data(fetched_at);
```

### 2.5 取得統計（オプション）

#### fetch_stats
データ取得の統計情報

```sql
CREATE TABLE fetch_stats (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    config_id UUID NOT NULL REFERENCES user_fetch_configs(id),
    fetched_at TIMESTAMP DEFAULT NOW(),
    items_found INTEGER DEFAULT 0,
    items_saved INTEGER DEFAULT 0,
    items_skipped INTEGER DEFAULT 0,
    error TEXT,
    duration_ms INTEGER
);
```

## 3. データソース固有設定の構造

各データソースごとに専用テーブルで設定を管理します。

### Reddit設定の例
- **subreddit**: 取得対象のサブレディット
- **sort_by**: ソート方法（hot, new, top, rising）
- **time_filter**: 期間フィルター（hour, day, week, month, year, all）
- **limit_count**: 取得件数制限
- **keywords**: キーワード配列

### YouTube設定の例
- **channel_id**: 特定チャンネルID
- **playlist_id**: 特定プレイリストID
- **keywords**: 検索キーワード配列
- **max_results**: 最大取得件数
- **order_by**: ソート方法（relevance, date, viewCount, rating, title）
- **published_after**: 指定日時以降の動画のみ取得

## 4. Row Level Security (RLS)

### users
```sql
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- ユーザーは自分の情報のみアクセス可能
CREATE POLICY "Users can view own profile" ON users
  FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON users
  FOR UPDATE USING (auth.uid() = id);
```

### user_fetch_configs
```sql
ALTER TABLE user_fetch_configs ENABLE ROW LEVEL SECURITY;

-- 自分の設定のみアクセス可能
CREATE POLICY "Users can view own configs" ON user_fetch_configs
  FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own configs" ON user_fetch_configs
  FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own configs" ON user_fetch_configs
  FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own configs" ON user_fetch_configs
  FOR DELETE USING (auth.uid() = user_id);
```

### reddit_fetch_configs
```sql
ALTER TABLE reddit_fetch_configs ENABLE ROW LEVEL SECURITY;

-- 自分の設定に紐づくReddit設定のみアクセス可能
CREATE POLICY "Users can view own reddit configs" ON reddit_fetch_configs
  FOR SELECT USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = reddit_fetch_configs.user_fetch_config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );

CREATE POLICY "Users can insert own reddit configs" ON reddit_fetch_configs
  FOR INSERT WITH CHECK (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = reddit_fetch_configs.user_fetch_config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );

CREATE POLICY "Users can update own reddit configs" ON reddit_fetch_configs
  FOR UPDATE USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = reddit_fetch_configs.user_fetch_config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );

CREATE POLICY "Users can delete own reddit configs" ON reddit_fetch_configs
  FOR DELETE USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = reddit_fetch_configs.user_fetch_config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );
```

### youtube_fetch_configs
```sql
ALTER TABLE youtube_fetch_configs ENABLE ROW LEVEL SECURITY;

-- 自分の設定に紐づくYouTube設定のみアクセス可能
CREATE POLICY "Users can view own youtube configs" ON youtube_fetch_configs
  FOR SELECT USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = youtube_fetch_configs.user_fetch_config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );

CREATE POLICY "Users can insert own youtube configs" ON youtube_fetch_configs
  FOR INSERT WITH CHECK (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = youtube_fetch_configs.user_fetch_config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );

CREATE POLICY "Users can update own youtube configs" ON youtube_fetch_configs
  FOR UPDATE USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = youtube_fetch_configs.user_fetch_config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );

CREATE POLICY "Users can delete own youtube configs" ON youtube_fetch_configs
  FOR DELETE USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = youtube_fetch_configs.user_fetch_config_id
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

### fetch_stats
```sql
ALTER TABLE fetch_stats ENABLE ROW LEVEL SECURITY;

-- 自分の設定の統計のみ参照可能
CREATE POLICY "Users can view own stats" ON fetch_stats
  FOR SELECT USING (
    EXISTS (
      SELECT 1 FROM user_fetch_configs
      WHERE user_fetch_configs.id = fetch_stats.config_id
      AND user_fetch_configs.user_id = auth.uid()
    )
  );
```

## 5. 初期データ

### データソース
```sql
INSERT INTO data_sources (id, name, icon) VALUES
  ('reddit', 'Reddit', '🟠'),
  ('twitter', 'Twitter', '🐦'),
  ('youtube', 'YouTube', '📺'),
  ('hackernews', 'Hacker News', '🟧'),
  ('github', 'GitHub', '🐙');
```

## 6. 使用例

### ユーザーがRedditの検索条件を設定する場合

```sql
-- 1. ユーザーを作成（通常はSupabase Authで自動作成）
INSERT INTO users (id, name) 
VALUES ('123e4567-e89b-12d3-a456-426614174000', 'John Doe');

-- 2. 共通の取得設定を作成
INSERT INTO user_fetch_configs (id, user_id, name, data_source_id) 
VALUES (
  'aaa11111-1111-1111-1111-111111111111',
  '123e4567-e89b-12d3-a456-426614174000',
  'Tech Subreddits',
  'reddit'
);

-- 3. Reddit固有設定を作成
INSERT INTO reddit_fetch_configs (
  user_fetch_config_id,
  subreddit,
  sort_by,
  time_filter,
  limit_count,
  keywords
) VALUES (
  'aaa11111-1111-1111-1111-111111111111',
  'golang',
  'hot',
  'day',
  25,
  ARRAY['backend', 'api', 'microservices']
);
```

### ユーザーがYouTubeの検索条件を設定する場合

```sql
-- 1. 共通の取得設定を作成
INSERT INTO user_fetch_configs (id, user_id, name, data_source_id) 
VALUES (
  'bbb22222-2222-2222-2222-222222222222',
  '123e4567-e89b-12d3-a456-426614174000',
  'Go Programming Videos',
  'youtube'
);

-- 2. YouTube固有設定を作成
INSERT INTO youtube_fetch_configs (
  user_fetch_config_id,
  channel_id,
  keywords,
  max_results,
  order_by
) VALUES (
  'bbb22222-2222-2222-2222-222222222222',
  'UC_x5XG1OV2P6uZZ5FSM9Ttw',
  ARRAY['golang', 'tutorial', 'programming'],
  50,
  'relevance'
);
```

### バッチ処理でアクティブな設定を取得

```sql
-- Reddit設定を含む全データを取得
SELECT 
  ufc.id,
  ufc.user_id,
  ufc.name,
  ufc.data_source_id,
  ds.name as data_source_name,
  rfc.subreddit,
  rfc.sort_by,
  rfc.time_filter,
  rfc.limit_count,
  rfc.keywords
FROM user_fetch_configs ufc
JOIN data_sources ds ON ufc.data_source_id = ds.id
LEFT JOIN reddit_fetch_configs rfc ON ufc.id = rfc.user_fetch_config_id
WHERE ufc.is_active = TRUE
  AND ds.is_active = TRUE
  AND ufc.data_source_id = 'reddit';

-- YouTube設定を含む全データを取得
SELECT 
  ufc.id,
  ufc.user_id,
  ufc.name,
  ufc.data_source_id,
  ds.name as data_source_name,
  yfc.channel_id,
  yfc.playlist_id,
  yfc.keywords,
  yfc.max_results,
  yfc.order_by,
  yfc.published_after
FROM user_fetch_configs ufc
JOIN data_sources ds ON ufc.data_source_id = ds.id
LEFT JOIN youtube_fetch_configs yfc ON ufc.id = yfc.user_fetch_config_id
WHERE ufc.is_active = TRUE
  AND ds.is_active = TRUE
  AND ufc.data_source_id = 'youtube';
```

### 取得したデータの保存

```sql
-- Redditから取得したデータを保存
INSERT INTO fetched_data (
  config_id,
  source,
  title,
  content,
  url,
  author_name,
  source_item_id,
  published_at,
  tags,
  metadata
) VALUES (
  'config-uuid-here',
  'reddit',
  'Interesting Go Performance Tips',
  'Here are some tips for improving Go performance...',
  'https://reddit.com/r/golang/comments/abc123',
  'gopher123',
  'abc123',
  '2024-01-15 10:30:00',
  ARRAY['golang', 'performance'],
  '{
    "score": 142,
    "num_comments": 23,
    "subreddit": "golang",
    "is_self": true
  }'::jsonb
);
```

## 7. 今後の拡張性

新しいデータソースを追加する場合：

1. `data_sources` テーブルに新しい行を追加
2. 新しいデータソース固有の設定テーブルを作成（例：`twitter_fetch_configs`）
3. 対応するGoのdomain modelとrepositoryを実装
4. バッチ処理のusecaseにswitch文の分岐を追加
5. 対応するfetcherを実装

この設計により、データソースごとの特性を活かしながら、型安全で拡張しやすい構造を実現できます。

### 新しいデータソース追加例（Twitter）

```sql
-- Twitter固有設定テーブル
CREATE TABLE twitter_fetch_configs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_fetch_config_id UUID NOT NULL REFERENCES user_fetch_configs(id) ON DELETE CASCADE,
    username TEXT, -- 特定ユーザーのツイート取得
    hashtags TEXT[], -- ハッシュタグでの検索
    keywords TEXT[], -- キーワードでの検索
    exclude_retweets BOOLEAN DEFAULT FALSE,
    include_replies BOOLEAN DEFAULT FALSE,
    max_results INTEGER DEFAULT 100,
    created_at TIMESTAMP DEFAULT NOW()
);
```