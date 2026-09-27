CREATE TABLE dbo.Log(
    LogID bigint IDENTITY(1,1) NOT NULL,
    LoggedAt datetime2(3) NOT NULL CONSTRAINT DF_Log_Timestamp DEFAULT SYSUTCDATETIME(),
    RequestID uniqueidentifier NULL,
    Event varchar(50) NOT NULL,
    ClientIP varchar(45) NULL,
    UserAgent nvarchar(500) NULL,
    ServerNumber int NOT NULL,
    Version varchar(30) NULL,
    Method varchar(10) NULL,
    Endpoint nvarchar(500) NULL,
    StatusCode smallint NULL,
    DurationMs int NULL,
    CacheKey varchar(200) NULL,
    QueryName varchar(200) NULL,
    Term varchar(50) NULL,
    DataSize int NULL,
    Message nvarchar(1000) NULL,
    Details nvarchar(max) NULL,
    CONSTRAINT PK_Log PRIMARY KEY CLUSTERED(LogID)
);

CREATE INDEX IX_Log_Timestamp ON dbo.Log(LoggedAt);
CREATE INDEX IX_Log_RequestID ON dbo.Log(RequestID);
CREATE INDEX IX_Log_Event_Timestamp ON dbo.Log(Event, LoggedAt);
