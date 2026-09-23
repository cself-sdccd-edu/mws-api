USE [MWSAPI]
GO

/****** Object:  Table [dbo].[Cache]    Script Date: 9/23/2026 1:55:38 PM ******/
SET ANSI_NULLS ON
GO

SET QUOTED_IDENTIFIER ON
GO

CREATE TABLE [dbo].[Cache](
	[CacheKey] [varchar](200) NOT NULL,
	[Data] [varbinary](max) NULL,
	[UpdatedAt] [datetime2](3) NOT NULL,
	[RefreshStartedAt] [datetime2](3) NULL,
	[LastRefreshError] [nvarchar](1000) NULL,
 CONSTRAINT [PK_Cache] PRIMARY KEY CLUSTERED 
(
	[CacheKey] ASC
)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON, OPTIMIZE_FOR_SEQUENTIAL_KEY = OFF) ON [PRIMARY]
) ON [PRIMARY] TEXTIMAGE_ON [PRIMARY]
GO

