/*　拡張機能を生成　*/
CREATE EXTENSION "uuid-ossp";
CREATE EXTENSION "pgcrypto";

CREATE ROLE bkowner WITH LOGIN;
CREATE ROLE aploper WITH LOGIN;
CREATE ROLE cmnoper WITH LOGIN;
CREATE ROLE archons WITH LOGIN;

/*　データベースを生成　*/
DROP DATABASE abook;
CREATE DATABASE abook WITH OWNER = bkowner;

\c abook bkowner


CREATE OR REPLACE FUNCTION IsNotNull (pText TEXT) RETURNS BOOL AS $$
DECLARE
  bResult BOOL;
BEGIN
  IF pText <> '' THEN 
    bResult := TRUE;
  ELSE
    bResult := FALSE;
  END IF;
  RETURN bResult;
END;
$$ LANGUAGE plpgsql;



/*　認証局管理テーブル　*/
DROP TABLE MAuthoritys CASCADE;
CREATE TABLE MAuthoritys (
  AuthorityId   UUID NOT NULL,
  AuthorityName VARCHAR(64) NOT NULL,
  PRIMARY KEY (AuthorityId)
);
GRANT ALL ON MAuthoritys TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON MAuthoritys TO aploper;

INSERT INTO MAuthoritys (AuthorityId, AuthorityName) VALUES ('{68d14f7b-b5ee-401e-b338-15efad4f87c2}','Google');
INSERT INTO MAuthoritys (AuthorityId, AuthorityName) VALUES ('{7bba2244-a60c-445e-9c8a-ab4cdf4e2473}','Entra');


/*　アカウント情報テーブル　*/
DROP TABLE MAccounts CASCADE;
CREATE TABLE MAccounts (
  UniqueId      UUID NOT NULL DEFAULT gen_random_uuid(),
  DisplayName   VARCHAR(256) NOT NULL,
  MailAddress   VARCHAR(256) NOT NULL,
  SecretHashV   VARCHAR(256) NOT NULL,
  CreateStamp   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  PRIMARY KEY (UniqueId)
);
GRANT ALL ON MAccounts TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON MAccounts TO aploper;


/*　アカウント紐付け情報テーブル　*/
/* 紐付けているアカウントがどれであるかを記録する。 */
DROP TABLE TAccessLink CASCADE;
CREATE TABLE TAccessLink (
  UniqueId      UUID NOT NULL,
  AuthorityId   UUID NOT NULL,
  AccountId     VARCHAR(256) NOT NULL,
  CreateStamp   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  PRIMARY KEY (UniqueId, AuthorityId, AccountId),
  FOREIGN KEY (UniqueId) REFERENCES MAccounts (UniqueId) ON DELETE CASCADE,
  CHECK (TRIM(AccountId) <> '')
);
GRANT ALL ON TAccessLink TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON TAccessLink TO aploper;

INSERT INTO MAccounts (UniqueId, DisplayName, MailAddress) VALUES ('{38f79db6-0483-496f-b250-53839ec1a73f}', 'display', 'mail@mail.net');
DELETE FROM TAccessLink WHERE UniqueId = '{38f79db6-0483-496f-b250-53839ec1a73f}';
INSERT INTO TAccessLink (UniqueId, AuthorityId, AccountId) (SELECT '{38f79db6-0483-496f-b250-53839ec1a73f}', AuthorityId, 'rink@arteria-s.net' FROM MAuthoritys WHERE AuthorityName = 'Google');
INSERT INTO TAccessLink (UniqueId, AuthorityId, AccountId) (SELECT '{38f79db6-0483-496f-b250-53839ec1a73f}', AuthorityId, 'rink@0adrastea.onmicrosoft.com' FROM MAuthoritys WHERE AuthorityName = 'Entra');


/*　セッション属性情報テーブル　*/
/*　セッションキーはユニーク。紐付くユニークキーは、匿名状態（NULL）を許可する必要　*/
DROP TABLE TSessions CASCADE;
CREATE TABLE TSessions (
  SessionKey    VARCHAR(256) NOT NULL,
  UniqueId      UUID NULL,
  ExpireStamp   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  PRIMARY KEY (SessionKey),
  FOREIGN KEY (UniqueId) REFERENCES MAccounts (UniqueId)
);
GRANT ALL ON TSessions TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON TSessions TO aploper;


/*　アクセストークン保管テーブル　*/
/*　一つのセッションキー毎に外部アカウント別のアクセストークンを保存する。　*/
DROP TABLE TSTokens CASCADE;
CREATE TABLE TSTokens (
  SessionKey    VARCHAR(256) NOT NULL,
  AuthorityId   UUID NOT NULL,
  AccessToken   TEXT NOT NULL,
  RefreshToken  TEXT NOT NULL,
  Expiry        TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY (SessionKey, AuthorityId),
  FOREIGN KEY (SessionKey) REFERENCES TSessions (SessionKey) ON DELETE CASCADE,
  FOREIGN KEY (AuthorityId) REFERENCES MAuthoritys (AuthorityId)
);
GRANT ALL ON TSTokens TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON TSTokens TO aploper;

CREATE VIEW VSTokens AS SELECT SessionKey, AuthorityId, Expiry FROM TSTokens;
GRANT ALL ON VSTokens TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON VSTokens TO aploper;





/*　アクセストークン保管テーブル　*/
/*　一つのユニークキー毎に外部アカウント別のアクセストークンを保存する。　*/
DROP TABLE TATokens CASCADE;
CREATE TABLE TATokens (
  UniqueId      UUID NOT NULL,
  AuthorityId   UUID NOT NULL,
  AccessToken   TEXT NOT NULL,
  RefreshToken  TEXT NOT NULL,
  Expiry        TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY (UniqueId, AuthorityId),
  FOREIGN KEY (UniqueId) REFERENCES MAccounts (UniqueId) ON DELETE CASCADE,
  FOREIGN KEY (AuthorityId) REFERENCES MAuthoritys (AuthorityId)
);
GRANT ALL ON TATokens TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON TATokens TO aploper;

DROP VIEW VATokens;
CREATE VIEW VATokens AS SELECT MailAddress AS UniqueName, AuthorityName, Expiry, IsNotNull(RefreshToken) AS RefreshToken  FROM TATokens LEFT JOIN MAuthoritys ON TATokens.AuthorityId = MAuthoritys.AuthorityId LEFT JOIN MAccounts ON TATokens.UniqueId = MAccounts.UniqueId;
GRANT ALL ON VATokens TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON VATokens TO aploper;





/*　チャレンジ状態記録テーブル（状態の記録と検証）　*/
DROP TABLE TChallenges CASCADE;
CREATE TABLE TChallenges (
  SessionKey    VARCHAR(256) NOT NULL,
  ChallengeVal  VARCHAR(256) NOT NULL,
  EventStamp    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  PRIMARY KEY (SessionKey,ChallengeVal)
);
GRANT ALL ON TChallenges TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON TChallenges TO aploper;



/*　ログイン成功履歴テーブル　*/
/*　外部の認証サーバーからアクセスが許可された時にレコードに記録を作る。　*/
DROP TABLE HLogins CASCADE;
CREATE TABLE HLogins (
  EventID       UUID NOT NULL DEFAULT gen_random_uuid(),
  UniqueId      UUID NOT NULL,
  AuthorityId   UUID NOT NULL,
  AccountId     VARCHAR(256) NOT NULL,
  ClientIP      VARCHAR(256) NOT NULL,
  EventStamp    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  PRIMARY KEY (EventID),
  FOREIGN KEY (UniqueId) REFERENCES MAccounts (UniqueId)
);
GRANT ALL ON HLogins TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON HLogins TO aploper;


/*　イベントフック状態記録テーブル　*/
DROP TABLE TWebHooks CASCADE;
CREATE TABLE TWebHooks (
  UniqueId      UUID NOT NULL,
  AuthorityId   UUID NOT NULL,
  AccountId     VARCHAR(256) NOT NULL,
  ResourceId    VARCHAR(512) NOT NULL,
  SubscribeId   VARCHAR(512) NOT NULL,
  ClientState   TEXT NOT NULL DEFAULT '',
  ExpireStamp   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  LastSyncStamp TIMESTAMP,
  PRIMARY KEY (UniqueId, AuthorityId, AccountId, ResourceId, SubscribeId),
  FOREIGN KEY (UniqueId) REFERENCES MAccounts (UniqueId) ON DELETE CASCADE,
  FOREIGN KEY (UniqueId, AuthorityId) REFERENCES TATokens (UniqueId, AuthorityId) ON DELETE CASCADE
);
GRANT ALL ON TWebHooks TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON TWebHooks TO aploper;

DROP VIEW VWebHooks;
CREATE VIEW VWebHooks AS SELECT MailAddress AS UniqueName, AuthorityName, AccountId, ResourceId, SubscribeId, ClientState, ExpireStamp, LastSyncStamp FROM TWebHooks LEFT JOIN MAuthoritys ON TWebHooks.AuthorityId = MAuthoritys.AuthorityId LEFT JOIN MAccounts ON TWebHooks.UniqueId = MAccounts.UniqueId;
GRANT ALL ON VWebHooks TO cmnoper;
GRANT SELECT, INSERT, UPDATE, DELETE ON VWebHooks TO aploper;



/*　　*/
