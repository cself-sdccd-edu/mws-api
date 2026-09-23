CREATE OR ALTER PROCEDURE dbo.Cache_Get
    @CacheKey varchar(200)
AS
BEGIN
    SET NOCOUNT ON;

    SELECT CacheKey, Data, UpdatedAt, RefreshStartedAt, LastRefreshError
    FROM dbo.Cache
    WHERE CacheKey = @CacheKey;
END
GO

