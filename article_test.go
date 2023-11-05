package nntpclient

var sourceLines = []string{
	"Path: news.netfront.net!goblin1!goblin3!goblin.stu.neva.ru!gandalf.srv.welterde.de!eternal-september.org!reader02.eternal-september.org!snipe.eternal-september.org!.POSTED!not-for-mail",
	"From: !@!.invalid (=?ISO-8859-1?Q?=CF?=)",
	"Newsgroups: free.rocks",
	"Subject: Gordon",
	"Date: Thu, 22 Apr 2021 14:08:35 +0100",
	"Organization: =?ISO-8859-1?Q?=CF?=",
	"Lines: 4",
	"Message-ID: <1p81j8s.l0kmdc1b6stnjN%!@!.invalid>",
	"Injection-Info: snipe.eternal-september.org; posting-host=\"6ff876e5a51a3ca52a5611e89065dcb0\";",
	"\tlogging-data=\"22206\"; mail-complaints-to=\"abuse@eternal-september.org\";	posting-account=\"U2FsdGVkX1/t0zsaJa7+EqlzAYXEAb1X\"",
	"User-Agent: MacSOUP/2.8.6b1 (ed136d9b90) (Mac OS 10.14.6)",
	"Cancel-Lock: sha1:0PE2VO4two8MzpUt/KpLJRk5p2A=",
	"X-No-Archive: !!",
	"Xref: news.netfront.net free.rocks:1",
	"",
	"Rock on, Gordon.",
	"",
	"--",
	"fold, spindle, mutilate.",
}

//func Test_readHeaders(t *testing.T) {
//	headers, bodyStart := readHeaders(sourceLines)
//
//	assert.Equal(t, 13, len(headers))
//	assert.Equal(t, 15, bodyStart)
//}
