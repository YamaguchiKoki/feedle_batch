# CLAUDE.md - Feedle_batch

## プロジェクト概要
Feedle_batchは、複数のソーシャルメディアプラットフォーム（Twitter、YouTube、Instagram、Reddit、Hacker News etc...）から情報を収集し、Supabaseデータベースに保存するGoベースのバッチ処理システムです。

## Conversation Guidelines

- 常に日本語で会話する

## Development Philosophy

### Test-Driven Development (TDD)

- 原則としてテスト駆動開発（TDD）で進める
- 期待される入出力に基づき、まずテストを作成する
- 実装コードは書かず、テストのみを用意する
- テストを実行し、失敗を確認する
- テストが正しいことを確認できた段階でコミットする
- その後、テストをパスさせる実装を進める
- 実装中はテストを変更せず、コードを修正し続ける
- すべてのテストが通過するまで繰り返す


## データベーススキーマ
正規化された設計を採用し、検索条件を柔軟に管理：@docs/db.md

### 主要テーブル
1. **user_fetch_configs** - ユーザーの取得設定
2. **user_conditions** - 検索条件
3. **user_condition_items** - 配列型条件の値
4. **fetched_data** - 取得したデータ


### テスト実行の自動化
```toml
# .mise.toml
[tasks.test]
run = "go test -v ./..."
description = "Run all tests"

```

```bash
# 使用例
mise run test
```
