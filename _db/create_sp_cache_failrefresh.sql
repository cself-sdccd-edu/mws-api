CREATE OR ALTER PROCEDURE dbo.Cache_FailRefresh
    @CacheKey varchar(200),
    @Error nvarchar(1000)
AS
BEGIN
    SET NOCOUNT ON;

    UPDATE dbo.Cache
    SET RefreshStartedAt = NULL,
        LastRefreshError = @Error
    WHERE CacheKey = @CacheKey;
END
GO

