"use strict";

var d           = require("d")
  , pad         = require("es5-ext/number/#/pad")
  , date        = require("es5-ext/date/valid-date")
  , daysInMonth = require("es5-ext/date/#/days-in-month")
  , copy        = require("es5-ext/date/#/copy")
  , dfloor      = require("es5-ext/date/#/floor-day")
  , mfloor      = require("es5-ext/date/#/floor-month")
  , yfloor      = require("es5-ext/date/#/floor-year")
  , toInteger   = require("es5-ext/number/to-integer")
  , toPosInt    = require("es5-ext/number/to-pos-integer")
  , isValue     = require("es5-ext/object/is-value");

var abs = Math.abs, format, toPrimitive, getYear, Duration, getCalcData;

format = require("es5-ext/string/format-method")({
	y: function () { return String(abs(this.year)); },
	m: function () { return pad.call(abs(this.month), 2); },
	d: function () { return pad.call(abs(this.day), 2); },
	H: function () { return pad.call(abs(this.hour), 2); },
	M: function () { return pad.call(abs(this.minute), 2); },
	S: function () { return pad.call(abs(this.second), 2); },
	L: function () { return pad.call(abs(this.millisecond), 3); },

	ms: function () { return String(abs(this.months)); },
	ds: function () { return String(abs(this.days)); },
	Hs: function () { return String(abs(this.hours)); },
	Ms: function () { return String(abs(this.minutes)); },
	Ss: function () { return String(abs(this.seconds)); },
	Ls: function () { return String(abs(this.milliseconds)); },

	sign: function () { return this.to < this.from ? "-" : ""; }
});

getCalcData = function (duration) {
	return duration.to < duration.from
		? { to: duration.from, from: duration.to, sign: -1 }
		: { to: duration.to, from: duration.from, sign: 1 };
};

Duration = module.exports = function (from, to) {
	// Make it both constructor and factory
	if (!(this instanceof Duration)) return new Duration(from, to);

	this.from = date(from);
	this.to = isValue(to) ? date(to) : new Date();
};

Duration.prototype = Object.create(Object.prototype, {
	valueOf: d((toPrimitive = function () { return this.to - this.from; })),
	millisecond: d.gs(function () { return this.milliseconds % 1000; }),
	second: d.gs(function () { return this.seconds % 60; }),
	minute: d.gs(function () { return this.minutes % 60; }),
	hour: d.gs(function () { return this.hours % 24; }),
	day: d.gs(function () {
		var data = getCalcData(this);
		var toDays = data.to.getDate(), fromDays = data.from.getDate();
		var isToLater =
			data.to - dfloor.call(copy.call(data.to)) >=
			data.from - dfloor.call(copy.call(data.from));
		var result;
		if (toDays > fromDays) {
			result = toDays - fromDays;
			if (!isToLater) --result;
			return data.sign * result;
		}
		if (toDays === fromDays && isToLater) {
			return 0;
		}
		result = isToLater ? toDays : toDays - 1;
		result += daysInMonth.call(data.from) - data.from.getDate();
		return data.sign * result;
	}),
	month: d.gs(function () {
		var data = getCalcData(this);
		return (
			data.sign *
			(((12 - data.from.getMonth() + data.to.getMonth()) % 12) -
				(data.from - mfloor.call(copy.call(data.from)) >
					data.to - mfloor.call(copy.call(data.to))))
		);
	}),
	year: d.gs(
		(getYear = function () {
			var data = getCalcData(this);
			return (
				data.sign *
				(data.to.getFullYear() -
					data.from.getFullYear() -
					(data.from - yfloor.call(copy.call(data.from)) >
						data.to - yfloor.call(copy.call(data.to))))
			);
		})
	),

	milliseconds: d.gs(toPrimitive, null),
	seconds: d.gs(function () { return toInteger(this.valueOf() / 1000); }),
	minutes: d.gs(function () { return toInteger(this.valueOf() / (1000 * 60)); }),
	hours: d.gs(function () { return toInteger(this.valueOf() / (1000 * 60 * 60)); }),
	days: d.gs(function () { return toInteger(this.valueOf() / (1000 * 60 * 60 * 24)); }),
	months: d.gs(function () {
		var data = getCalcData(this);
		return (
			data.sign *
			((data.to.getFullYear() - data.from.getFullYear()) * 12 +
				data.to.getMonth() -
				data.from.getMonth() -
				(data.from - mfloor.call(copy.call(data.from)) >
					data.to - mfloor.call(copy.call(data.to))))
		);
	}),
	years: d.gs(getYear),

	_resolveSign: d(function (isNonZero) {
		if (!isNonZero) return "";
		return this.to < this.from ? "-" : "";
	}),
	_toStringDefaultDate: d(function (threshold, s, isNonZero) {
		if (!this.days && threshold < 0) return this._resolveSign(isNonZero) + s;
		if (threshold-- <= 0) s = abs((isNonZero = this.day)) + "d" + (s ? " " : "") + s;
		if (!this.months && threshold < 0) return this._resolveSign(isNonZero) + s;
		if (threshold-- <= 0) s = abs((isNonZero = this.month)) + "m" + (s ? " " : "") + s;
		if (this.years || threshold >= 0) {
			s = abs((isNonZero = this.year)) + "y" + (s ? " " : "") + s;
		}
		return this._resolveSign(isNonZero) + s;
	}),
	_toStringDefault: d(function (threshold) {
		var s = "", isNonZero;
		if (threshold-- <= 0) s += "." + pad.call(abs((isNonZero = this.millisecond)), 3);
		if (!this.seconds && threshold < 0) return this._resolveSign(isNonZero) + s;
		if (threshold-- <= 0) {
			isNonZero = this.second;
			s = (this.minutes ? pad.call(abs(isNonZero), 2) : abs(isNonZero)) + s;
		}
		if (!this.minutes && threshold < 0) return this._resolveSign(isNonZero) + s;
		if (threshold-- <= 0) {
			isNonZero = this.minute;
			s =
				(this.hours || s ? pad.call(abs(isNonZero), 2) : abs(isNonZero)) +
				(s ? ":" : "") +
				s;
		}
		if (!this.hours && threshold < 0) return this._resolveSign(isNonZero) + s;
		if (threshold-- <= 0) s = pad.call(abs((isNonZero = this.hour)), 2) + (s ? ":" : "") + s;
		return this._toStringDefaultDate(threshold, s, isNonZero);
	}),
	_toString1: d(function (threshold) {
		var tokens = [], isNonZero;
		if (threshold-- <= 0) tokens.unshift(abs((isNonZero = this.millisecond)) + "ms");
		if (!this.seconds && threshold < 0) return this._resolveSign(isNonZero) + tokens.join(" ");
		if (threshold-- <= 0) tokens.unshift(abs((isNonZero = this.second)) + "s");
		if (!this.minutes && threshold < 0) return this._resolveSign(isNonZero) + tokens.join(" ");
		if (threshold-- <= 0) tokens.unshift(abs((isNonZero = this.minute)) + "m");
		if (!this.hours && threshold < 0) return this._resolveSign(isNonZero) + tokens.join(" ");
		if (threshold-- <= 0) tokens.unshift(abs((isNonZero = this.hour)) + "h");
		if (!this.days && threshold < 0) return this._resolveSign(isNonZero) + tokens.join(" ");
		if (threshold-- <= 0) tokens.unshift(abs((isNonZero = this.day)) + "d");
		if (!this.months && threshold < 0) return this._resolveSign(isNonZero) + tokens.join(" ");
		if (threshold-- <= 0) tokens.unshift(abs((isNonZero = this.month)) + "m");
		if (!this.years && threshold < 0) return this._resolveSign(isNonZero) + tokens.join(" ");
		tokens.unshift(abs((isNonZero = this.year)) + "y");
		return this._resolveSign(isNonZero) + tokens.join(" ");
	}),
	toString: d(function (pattern/*, threshold*/) {
		var threshold;
		if (!isValue(pattern)) pattern = 0;
		if (isNaN(pattern)) return format.call(this, pattern);
		pattern = Number(pattern);
		threshold = toPosInt(arguments[1]);
		if (pattern === 1) return this._toString1(threshold);
		return this._toStringDefault(threshold);
	})
});
