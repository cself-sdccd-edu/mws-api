-- NOTE
-- Use a random password instead of ... for creating the login
-- Depending on your environment, add the correct database name instead of API_DB_NAME
USE [master]
GO

CREATE LOGIN [api_user] WITH PASSWORD=N'...', DEFAULT_DATABASE=[MWSAPI], DEFAULT_LANGUAGE=[us_english], CHECK_EXPIRATION=OFF, CHECK_POLICY=OFF
GO

USE [API_DB_NAME]
GO

CREATE USER [api_user] FOR LOGIN [api_user] WITH DEFAULT_SCHEMA=[dbo]
GO

GRANT EXECUTE ON OBJECT::dbo.Cache_Get TO [api_user]
GO

GRANT EXECUTE ON OBJECT::dbo.Cache_TryStartRefresh TO [api_user]
GO

GRANT EXECUTE ON OBJECT::dbo.Cache_Save TO [api_user]
GO

GRANT EXECUTE ON OBJECT::dbo.Cache_FailRefresh TO [api_user]
GO

GRANT EXECUTE ON OBJECT::dbo.Log_Insert TO [api_user];
GO
