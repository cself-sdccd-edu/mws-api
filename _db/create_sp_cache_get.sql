CREATE OR ALTER PROCEDURE dbo.Cache_Get
    @CacheKey varchar(200)
AS
BEGIN
    SET NOCOUNT ON;

    SELECT CacheKey, Data, UpdatedAt, RefreshStartedAt, LastRefreshError, CASE WHEN Data IS NULL THEN 0 ELSE 1 END AS HasData
    FROM dbo.Cache
    WHERE CacheKey = @CacheKey;

END
GO

