CREATE OR ALTER PROCEDURE dbo.Cache_Save
    @CacheKey varchar(200),
    @Data varbinary(max)
AS
BEGIN
    SET NOCOUNT ON;
    SET XACT_ABORT ON;

    UPDATE dbo.Cache
    SET Data = @Data,
        UpdatedAt = SYSUTCDATETIME(),
        RefreshStartedAt = NULL,
        LastRefreshError = NULL
    WHERE CacheKey = @CacheKey;
END
GO

