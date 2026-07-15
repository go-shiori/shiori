package webserver

// type void struct{}

// type TypeBookmarkDtoUrl string

// type typeMockDB struct {
// 	model.DB
// 	MockSaveBookmarksFunc func(ctx context.Context, create bool, bookmarks ...model.BookmarkDTO) ([]model.BookmarkDTO, error)
// 	MockSaveBookmarksData map[TypeBookmarkDtoUrl]void
// }

// func (m *typeMockDB) SaveBookmarks(ctx context.Context, create bool, bookmarks ...model.BookmarkDTO) ([]model.BookmarkDTO, error) {
// 	if m.MockSaveBookmarksFunc != nil {
// 		return m.MockSaveBookmarksFunc(ctx, create, bookmarks...)
// 	}

// 	return bookmarks, nil
// }

// func TestApiInsertBookmarkDuplicate(t *testing.T) {
// 	var err error

// 	var mockDb *typeMockDB
// 	var mockHandler Handler

// 	var urlUrl *url.URL

// 	var mockHttpResponseWriter1 http.ResponseWriter
// 	var mockHttpResponseWriter2 http.ResponseWriter

// 	/*** * * ***/

// 	mockDb = &typeMockDB{
// 		MockSaveBookmarksFunc: func(
// 			ctx context.Context,
// 			create bool,
// 			bookmarks ...model.BookmarkDTO,
// 		) (
// 			[]model.BookmarkDTO,
// 			error,
// 		) {
// 			var err error
// 			var resBookmarkDtos []model.BookmarkDTO

// 			/*** * * ***/

// 			for _, bookmark := range bookmarks {
// 				var bookmarkUrl TypeBookmarkDtoUrl

// 				/*** * * ***/

// 				bookmarkUrl = TypeBookmarkDtoUrl(bookmark.URL)

// 				/*** * * ***/

// 				if _, exists := mockDb.MockSaveBookmarksData[bookmarkUrl]; exists {
// 					mockDb.MockSaveBookmarksData[bookmarkUrl] = void{}

// 					resBookmarkDtos = append(resBookmarkDtos, bookmark)
// 				} else {
// 					err = errors.New("")

// 					return resBookmarkDtos, err
// 				}
// 			}

// 			/*** * * ***/

// 			return resBookmarkDtos, err
// 		},
// 	}

// 	mockHandler = Handler{DB: mockDb}

// 	urlUrl, err = url.Parse("https://www.example.com")
// 	assert.True(t, err == nil) // instead of panic

// 	/*** * * ***/

// 	// 1st insert
// 	mockHandler.ApiInsertBookmark(
// 		mockHttpResponseWriter1,
// 		&http.Request{
// 			URL: urlUrl,
// 		},
// 		httprouter.Params{},
// 	)
// 	// assert.Equal(t, http.StatusOK, mockHttpResponseWriter1) // todo; its obviously wrong

// 	// 2nd save, duplicate
// 	mockHandler.ApiInsertBookmark(
// 		mockHttpResponseWriter2,
// 		&http.Request{
// 			URL: urlUrl,
// 		},
// 		httprouter.Params{},
// 	)
// 	// assert // todo; its obviously wrong
// }
