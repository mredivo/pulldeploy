package logging

type Writer struct {
	writerName string
	logLevel   severity // The severity threshold for allowing output for this writer
	l          *Logger  // The underlying Logger
}

func (lw *Writer) Debug(format string, a ...interface{}) {
	if kDEBUG < lw.logLevel {
		return
	}
	lw.l.debug(lw.writerName, format, a...)
}

func (lw *Writer) Info(format string, a ...interface{}) {
	if kINFO < lw.logLevel {
		return
	}
	lw.l.info(lw.writerName, format, a...)
}

func (lw *Writer) Warn(format string, a ...interface{}) {
	if kWARN < lw.logLevel {
		return
	}
	lw.l.warn(lw.writerName, format, a...)
}

func (lw *Writer) Error(format string, a ...interface{}) {
	if kERROR < lw.logLevel {
		return
	}
	lw.l.error(lw.writerName, format, a...)
}

func (lw *Writer) Fatal(format string, a ...interface{}) {
	lw.l.fatal(lw.writerName, format, a...)
}
