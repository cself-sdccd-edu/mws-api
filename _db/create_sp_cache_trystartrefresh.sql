CREATE OR ALTER PROCEDURE dbo.Cache_TryStartRefresh
    @CacheKey varchar(200),
    @LeaseSeconds int
AS
BEGIN
    SET NOCOUNT ON;
    SET XACT_ABORT ON;

    DECLARE @Now datetime2(3) = SYSUTCDATETIME();
    DECLARE @LeaseExpires datetime2(3) = DATEADD(SECOND, -@LeaseSeconds, @Now);
    DECLARE @Claimed bit = 0;

    BEGIN TRANSACTION;

    IF EXISTS (
        SELECT 1
        FROM dbo.Cache WITH (UPDLOCK, HOLDLOCK)
        WHERE CacheKey = @CacheKey
    )
    BEGIN
        UPDATE dbo.Cache
        SET RefreshStartedAt = @Now
        WHERE CacheKey = @CacheKey
          AND (RefreshStartedAt IS NULL OR RefreshStartedAt < @LeaseExpires);

        IF @@ROWCOUNT = 1
            SET @Claimed = 1;
    END
    ELSE
    BEGIN
        INSERT INTO dbo.Cache (CacheKey, Data, UpdatedAt, RefreshStartedAt, LastRefreshError)
        VALUES (@CacheKey, 0x, @Now, @Now, NULL);

        SET @Claimed = 1;
    END

    COMMIT TRANSACTION;

    SELECT @Claimed AS Claimed;
END
GO

