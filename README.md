# saml-proxy

SAML認証を行うproxyサーバー

EnvoyのAuthzを利用して認証済みかどうかをCookieから判定する。

認証されていない場合はIDPにリダイレクトし、ログインすることで認証される。

ログインURLに関しては`/saml/acs`となっている。

セッションの管理についてはCookieもしくはRedisサーバーの2種類の管理方法を提供している。

## Configuration

[sample](./saml.exampl.yaml)

| Key          | Overview                                |
| :----------- | :-------------------------------------- |
| metadata_url | IDPが用意しているMetadataURLを記述      |
| x509_cert    | SSL証明書ファイルパスを記述             |
| x509_key     | 秘密鍵のファイルパスを記述              |
| root_url     | リバースプロキシを含めたルートURLを記述 |
