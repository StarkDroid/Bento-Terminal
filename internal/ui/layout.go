package ui

type box struct {
	W int
	H int
}

type layout struct {
	Main    box
	Files   box
	History box
	Build   box
	Status  box
	OK      bool
}

func computeLayout(width, height int) layout {
	const (
		statusH = 1
		gap     = 1
	)
	var l layout
	if width < 72 || height < 20 {
		return l
	}
	sideW := width * 34 / 100
	if sideW < 28 {
		sideW = 28
	}
	if sideW > 40 {
		sideW = 40
	}
	mainW := width - gap - sideW
	if mainW < 32 {
		return l
	}
	inner := height - statusH - gap
	buildH := inner * 34 / 100
	if buildH < 8 {
		buildH = 8
	}
	if buildH > 16 {
		buildH = 16
	}
	topH := inner - buildH
	if topH < 9 {
		buildH -= 9 - topH
		topH = 9
	}
	if buildH < 8 {
		return l
	}
	filesH := topH * 55 / 100
	if filesH < 5 {
		filesH = 5
	}
	histH := topH - gap - filesH
	if histH < 4 {
		filesH -= 4 - histH
		histH = 4
	}
	if filesH < 4 || histH < 4 {
		return l
	}
	l.OK = true
	l.Main = box{W: mainW, H: topH}
	l.Files = box{W: sideW, H: filesH}
	l.History = box{W: sideW, H: histH}
	l.Build = box{W: width, H: buildH}
	l.Status = box{W: width, H: statusH}
	return l
}

func buildParts(panelH int) (listH, outH int) {
	body := panelH - 3
	if body <= 0 {
		return 0, 0
	}
	if body < 4 {
		return body, 0
	}
	listH = body / 2
	if listH < 2 {
		listH = 2
	}
	outH = body - 1 - listH
	if outH < 1 {
		outH = 1
		listH = body - 2
	}
	if listH < 1 {
		listH = 1
	}
	return listH, outH
}
