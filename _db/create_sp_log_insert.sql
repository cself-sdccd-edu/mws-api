CREATE OR ALTER   PROCEDURE [dbo].[Log_Insert]
    @RequestID uniqueidentifier = NULL,
    @Event varchar(50),
    @ClientIP varchar(45) = NULL,
    @UserAgent nvarchar(500) = NULL,
    @ServerNumber int,
    @Version varchar(30) = NULL,
    @Method varchar(10) = NULL,
    @Endpoint nvarchar(500) = NULL,
    @StatusCode smallint = NULL,
    @DurationMs int = NULL,
    @CacheKey varchar(200) = NULL,
    @QueryName varchar(200) = NULL,
    @Term varchar(50) = NULL,
    @DataSize int = NULL,
    @Message nvarchar(1000) = NULL,
    @Details nvarchar(max) = NULL
AS
BEGIN
    SET NOCOUNT ON;

    INSERT dbo.Log(
        RequestID,
        Event,
        ClientIP,
        UserAgent,
        ServerNumber,
        Version,
        Method,
        Endpoint,
        StatusCode,
        DurationMs,
        CacheKey,
        QueryName,
        Term,
        DataSize,
        Message,
        Details
    )
    VALUES(
        @RequestID,
        @Event,
        @ClientIP,
        @UserAgent,
        @ServerNumber,
        @Version,
        @Method,
        @Endpoint,
        @StatusCode,
        @DurationMs,
        @CacheKey,
        @QueryName,
        @Term,
        @DataSize,
        @Message,
        @Details
    );
END;
GO
