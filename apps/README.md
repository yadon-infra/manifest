# App of Apps Pattern

このディレクトリはArgoCDのApp of Appsパターンを使用して、すべてのアプリケーションを管理します。

## 構造

```
apps/
├── root-app.yaml          # ルートApplication (これをArgoCDに適用)
├── helm/                  # Helmチャートベースのアプリケーション
│   ├── harbor.yaml
│   ├── longhorn.yaml
│   ├── plane.yaml
│   ├── traefik.yaml
│   ├── atlantis.yaml
│   └── prometheus.yaml
└── kustomize/             # Kustomizeベースのアプリケーション
    ├── argocd.yaml
    ├── cilium.yaml
    ├── cloudflared.yaml
    ├── grafana.yaml
    ├── minio.yaml
    ├── minio-operator.yaml
    ├── healthserver.yaml
    └── longhorn-resources.yaml
```

## デプロイ方法

### 初回セットアップ

1. ArgoCDがすでにインストールされている場合、ルートApplicationを適用:

```bash
kubectl apply -f apps/root-app.yaml
```

2. すべてのアプリケーションが自動的に作成され、同期されます。

### Helm対応について

Helmチャートを使用するアプリケーションは、`sources`フィールドを使用して複数のソースを指定しています:

- 1つ目のソース: Helmチャートリポジトリ
- 2つ目のソース: values.yamlを含むGitリポジトリ

例 (harbor):
```yaml
sources:
  - repoURL: https://helm.goharbor.io
    chart: harbor
    targetRevision: 1.16.1
    helm:
      releaseName: harbor
      valueFiles:
        - $values/harbor/values.yaml
  - repoURL: https://github.com/yadon-infra/manifest
    targetRevision: main
    ref: values
```

### アプリケーションの追加

新しいアプリケーションを追加するには:

1. `apps/helm/` または `apps/kustomize/` に新しいApplicationマニフェストを作成
2. コミットしてプッシュ
3. root-appが自動的に新しいApplicationを検出して作成

### アプリケーションの削除

1. `apps/` ディレクトリから対応するマニフェストを削除
2. コミットしてプッシュ
3. root-appが自動的に削除を検出してアプリケーションを削除

## 特徴

- **自動同期**: すべてのアプリケーションは自動同期が有効
- **自己修復**: 手動変更は自動的に元に戻されます
- **Prune**: 削除されたリソースは自動的にクリーンアップ
- **ServerSideApply**: 大きなアノテーションの問題を回避
- **StatefulSet VolumeClaimTemplates**: 無視設定により、不要な再作成を防止

## 移行前との違い

### ApplicationSetからの変更点

- ApplicationSetは単一のマニフェストで複数のApplicationを生成
- App of Appsは各Applicationを個別に定義
- より細かい制御と設定が可能
- Helm対応が容易

### 利点

1. **Helm対応**: Helmチャートとvalues.yamlの統合が容易
2. **柔軟性**: 各アプリケーションごとに異なる設定が可能
3. **可読性**: 各アプリケーションの設定が明確
4. **デバッグ**: 問題の特定と修正が容易
