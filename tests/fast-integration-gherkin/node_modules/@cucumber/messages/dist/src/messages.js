"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.TestStepFinished = exports.TestRunStarted = exports.TestRunFinished = exports.TestCaseStarted = exports.TestCaseFinished = exports.TestStep = exports.StepMatchArgumentsList = exports.StepMatchArgument = exports.Group = exports.TestCase = exports.StepDefinitionPattern = exports.StepDefinition = exports.JavaStackTraceElement = exports.JavaMethod = exports.SourceReference = exports.Source = exports.PickleTag = exports.PickleTableRow = exports.PickleTableCell = exports.PickleTable = exports.PickleStepArgument = exports.PickleStep = exports.PickleDocString = exports.Pickle = exports.ParseError = exports.ParameterType = exports.Product = exports.Git = exports.Ci = exports.Meta = exports.Location = exports.Hook = exports.Tag = exports.TableRow = exports.TableCell = exports.Step = exports.Scenario = exports.RuleChild = exports.Rule = exports.FeatureChild = exports.Feature = exports.Examples = exports.DocString = exports.DataTable = exports.Comment = exports.Background = exports.GherkinDocument = exports.Envelope = exports.Duration = exports.Attachment = void 0;
exports.TestStepResultStatus = exports.StepDefinitionPatternType = exports.SourceMediaType = exports.AttachmentContentEncoding = exports.UndefinedParameterType = exports.Timestamp = exports.TestStepStarted = exports.TestStepResult = void 0;
const class_transformer_1 = require("class-transformer");
require("reflect-metadata");
class Attachment {
    constructor() {
        this.body = '';
        this.contentEncoding = AttachmentContentEncoding.IDENTITY;
        this.mediaType = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Source)
], Attachment.prototype, "source", void 0);
exports.Attachment = Attachment;
class Duration {
    constructor() {
        this.seconds = 0;
        this.nanos = 0;
    }
}
exports.Duration = Duration;
class Envelope {
}
__decorate([
    class_transformer_1.Type(() => Attachment)
], Envelope.prototype, "attachment", void 0);
__decorate([
    class_transformer_1.Type(() => GherkinDocument)
], Envelope.prototype, "gherkinDocument", void 0);
__decorate([
    class_transformer_1.Type(() => Hook)
], Envelope.prototype, "hook", void 0);
__decorate([
    class_transformer_1.Type(() => Meta)
], Envelope.prototype, "meta", void 0);
__decorate([
    class_transformer_1.Type(() => ParameterType)
], Envelope.prototype, "parameterType", void 0);
__decorate([
    class_transformer_1.Type(() => ParseError)
], Envelope.prototype, "parseError", void 0);
__decorate([
    class_transformer_1.Type(() => Pickle)
], Envelope.prototype, "pickle", void 0);
__decorate([
    class_transformer_1.Type(() => Source)
], Envelope.prototype, "source", void 0);
__decorate([
    class_transformer_1.Type(() => StepDefinition)
], Envelope.prototype, "stepDefinition", void 0);
__decorate([
    class_transformer_1.Type(() => TestCase)
], Envelope.prototype, "testCase", void 0);
__decorate([
    class_transformer_1.Type(() => TestCaseFinished)
], Envelope.prototype, "testCaseFinished", void 0);
__decorate([
    class_transformer_1.Type(() => TestCaseStarted)
], Envelope.prototype, "testCaseStarted", void 0);
__decorate([
    class_transformer_1.Type(() => TestRunFinished)
], Envelope.prototype, "testRunFinished", void 0);
__decorate([
    class_transformer_1.Type(() => TestRunStarted)
], Envelope.prototype, "testRunStarted", void 0);
__decorate([
    class_transformer_1.Type(() => TestStepFinished)
], Envelope.prototype, "testStepFinished", void 0);
__decorate([
    class_transformer_1.Type(() => TestStepStarted)
], Envelope.prototype, "testStepStarted", void 0);
__decorate([
    class_transformer_1.Type(() => UndefinedParameterType)
], Envelope.prototype, "undefinedParameterType", void 0);
exports.Envelope = Envelope;
class GherkinDocument {
    constructor() {
        this.comments = [];
    }
}
__decorate([
    class_transformer_1.Type(() => Feature)
], GherkinDocument.prototype, "feature", void 0);
__decorate([
    class_transformer_1.Type(() => Comment)
], GherkinDocument.prototype, "comments", void 0);
exports.GherkinDocument = GherkinDocument;
class Background {
    constructor() {
        this.location = new Location();
        this.keyword = '';
        this.name = '';
        this.description = '';
        this.steps = [];
        this.id = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], Background.prototype, "location", void 0);
__decorate([
    class_transformer_1.Type(() => Step)
], Background.prototype, "steps", void 0);
exports.Background = Background;
class Comment {
    constructor() {
        this.location = new Location();
        this.text = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], Comment.prototype, "location", void 0);
exports.Comment = Comment;
class DataTable {
    constructor() {
        this.location = new Location();
        this.rows = [];
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], DataTable.prototype, "location", void 0);
__decorate([
    class_transformer_1.Type(() => TableRow)
], DataTable.prototype, "rows", void 0);
exports.DataTable = DataTable;
class DocString {
    constructor() {
        this.location = new Location();
        this.content = '';
        this.delimiter = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], DocString.prototype, "location", void 0);
exports.DocString = DocString;
class Examples {
    constructor() {
        this.location = new Location();
        this.tags = [];
        this.keyword = '';
        this.name = '';
        this.description = '';
        this.tableBody = [];
        this.id = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], Examples.prototype, "location", void 0);
__decorate([
    class_transformer_1.Type(() => Tag)
], Examples.prototype, "tags", void 0);
__decorate([
    class_transformer_1.Type(() => TableRow)
], Examples.prototype, "tableHeader", void 0);
__decorate([
    class_transformer_1.Type(() => TableRow)
], Examples.prototype, "tableBody", void 0);
exports.Examples = Examples;
class Feature {
    constructor() {
        this.location = new Location();
        this.tags = [];
        this.language = '';
        this.keyword = '';
        this.name = '';
        this.description = '';
        this.children = [];
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], Feature.prototype, "location", void 0);
__decorate([
    class_transformer_1.Type(() => Tag)
], Feature.prototype, "tags", void 0);
__decorate([
    class_transformer_1.Type(() => FeatureChild)
], Feature.prototype, "children", void 0);
exports.Feature = Feature;
class FeatureChild {
}
__decorate([
    class_transformer_1.Type(() => Rule)
], FeatureChild.prototype, "rule", void 0);
__decorate([
    class_transformer_1.Type(() => Background)
], FeatureChild.prototype, "background", void 0);
__decorate([
    class_transformer_1.Type(() => Scenario)
], FeatureChild.prototype, "scenario", void 0);
exports.FeatureChild = FeatureChild;
class Rule {
    constructor() {
        this.location = new Location();
        this.tags = [];
        this.keyword = '';
        this.name = '';
        this.description = '';
        this.children = [];
        this.id = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], Rule.prototype, "location", void 0);
__decorate([
    class_transformer_1.Type(() => Tag)
], Rule.prototype, "tags", void 0);
__decorate([
    class_transformer_1.Type(() => RuleChild)
], Rule.prototype, "children", void 0);
exports.Rule = Rule;
class RuleChild {
}
__decorate([
    class_transformer_1.Type(() => Background)
], RuleChild.prototype, "background", void 0);
__decorate([
    class_transformer_1.Type(() => Scenario)
], RuleChild.prototype, "scenario", void 0);
exports.RuleChild = RuleChild;
class Scenario {
    constructor() {
        this.location = new Location();
        this.tags = [];
        this.keyword = '';
        this.name = '';
        this.description = '';
        this.steps = [];
        this.examples = [];
        this.id = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], Scenario.prototype, "location", void 0);
__decorate([
    class_transformer_1.Type(() => Tag)
], Scenario.prototype, "tags", void 0);
__decorate([
    class_transformer_1.Type(() => Step)
], Scenario.prototype, "steps", void 0);
__decorate([
    class_transformer_1.Type(() => Examples)
], Scenario.prototype, "examples", void 0);
exports.Scenario = Scenario;
class Step {
    constructor() {
        this.location = new Location();
        this.keyword = '';
        this.text = '';
        this.id = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], Step.prototype, "location", void 0);
__decorate([
    class_transformer_1.Type(() => DocString)
], Step.prototype, "docString", void 0);
__decorate([
    class_transformer_1.Type(() => DataTable)
], Step.prototype, "dataTable", void 0);
exports.Step = Step;
class TableCell {
    constructor() {
        this.location = new Location();
        this.value = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], TableCell.prototype, "location", void 0);
exports.TableCell = TableCell;
class TableRow {
    constructor() {
        this.location = new Location();
        this.cells = [];
        this.id = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], TableRow.prototype, "location", void 0);
__decorate([
    class_transformer_1.Type(() => TableCell)
], TableRow.prototype, "cells", void 0);
exports.TableRow = TableRow;
class Tag {
    constructor() {
        this.location = new Location();
        this.name = '';
        this.id = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Location)
], Tag.prototype, "location", void 0);
exports.Tag = Tag;
class Hook {
    constructor() {
        this.id = '';
        this.sourceReference = new SourceReference();
    }
}
__decorate([
    class_transformer_1.Type(() => SourceReference)
], Hook.prototype, "sourceReference", void 0);
exports.Hook = Hook;
class Location {
    constructor() {
        this.line = 0;
    }
}
exports.Location = Location;
class Meta {
    constructor() {
        this.protocolVersion = '';
        this.implementation = new Product();
        this.runtime = new Product();
        this.os = new Product();
        this.cpu = new Product();
    }
}
__decorate([
    class_transformer_1.Type(() => Product)
], Meta.prototype, "implementation", void 0);
__decorate([
    class_transformer_1.Type(() => Product)
], Meta.prototype, "runtime", void 0);
__decorate([
    class_transformer_1.Type(() => Product)
], Meta.prototype, "os", void 0);
__decorate([
    class_transformer_1.Type(() => Product)
], Meta.prototype, "cpu", void 0);
__decorate([
    class_transformer_1.Type(() => Ci)
], Meta.prototype, "ci", void 0);
exports.Meta = Meta;
class Ci {
    constructor() {
        this.name = '';
    }
}
__decorate([
    class_transformer_1.Type(() => Git)
], Ci.prototype, "git", void 0);
exports.Ci = Ci;
class Git {
    constructor() {
        this.remote = '';
        this.revision = '';
    }
}
exports.Git = Git;
class Product {
    constructor() {
        this.name = '';
    }
}
exports.Product = Product;
class ParameterType {
    constructor() {
        this.name = '';
        this.regularExpressions = [];
        this.preferForRegularExpressionMatch = false;
        this.useForSnippets = false;
        this.id = '';
    }
}
exports.ParameterType = ParameterType;
class ParseError {
    constructor() {
        this.source = new SourceReference();
        this.message = '';
    }
}
__decorate([
    class_transformer_1.Type(() => SourceReference)
], ParseError.prototype, "source", void 0);
exports.ParseError = ParseError;
class Pickle {
    constructor() {
        this.id = '';
        this.uri = '';
        this.name = '';
        this.language = '';
        this.steps = [];
        this.tags = [];
        this.astNodeIds = [];
    }
}
__decorate([
    class_transformer_1.Type(() => PickleStep)
], Pickle.prototype, "steps", void 0);
__decorate([
    class_transformer_1.Type(() => PickleTag)
], Pickle.prototype, "tags", void 0);
exports.Pickle = Pickle;
class PickleDocString {
    constructor() {
        this.content = '';
    }
}
exports.PickleDocString = PickleDocString;
class PickleStep {
    constructor() {
        this.astNodeIds = [];
        this.id = '';
        this.text = '';
    }
}
__decorate([
    class_transformer_1.Type(() => PickleStepArgument)
], PickleStep.prototype, "argument", void 0);
exports.PickleStep = PickleStep;
class PickleStepArgument {
}
__decorate([
    class_transformer_1.Type(() => PickleDocString)
], PickleStepArgument.prototype, "docString", void 0);
__decorate([
    class_transformer_1.Type(() => PickleTable)
], PickleStepArgument.prototype, "dataTable", void 0);
exports.PickleStepArgument = PickleStepArgument;
class PickleTable {
    constructor() {
        this.rows = [];
    }
}
__decorate([
    class_transformer_1.Type(() => PickleTableRow)
], PickleTable.prototype, "rows", void 0);
exports.PickleTable = PickleTable;
class PickleTableCell {
    constructor() {
        this.value = '';
    }
}
exports.PickleTableCell = PickleTableCell;
class PickleTableRow {
    constructor() {
        this.cells = [];
    }
}
__decorate([
    class_transformer_1.Type(() => PickleTableCell)
], PickleTableRow.prototype, "cells", void 0);
exports.PickleTableRow = PickleTableRow;
class PickleTag {
    constructor() {
        this.name = '';
        this.astNodeId = '';
    }
}
exports.PickleTag = PickleTag;
class Source {
    constructor() {
        this.uri = '';
        this.data = '';
        this.mediaType = SourceMediaType.TEXT_X_CUCUMBER_GHERKIN_PLAIN;
    }
}
exports.Source = Source;
class SourceReference {
}
__decorate([
    class_transformer_1.Type(() => JavaMethod)
], SourceReference.prototype, "javaMethod", void 0);
__decorate([
    class_transformer_1.Type(() => JavaStackTraceElement)
], SourceReference.prototype, "javaStackTraceElement", void 0);
__decorate([
    class_transformer_1.Type(() => Location)
], SourceReference.prototype, "location", void 0);
exports.SourceReference = SourceReference;
class JavaMethod {
    constructor() {
        this.className = '';
        this.methodName = '';
        this.methodParameterTypes = [];
    }
}
exports.JavaMethod = JavaMethod;
class JavaStackTraceElement {
    constructor() {
        this.className = '';
        this.fileName = '';
        this.methodName = '';
    }
}
exports.JavaStackTraceElement = JavaStackTraceElement;
class StepDefinition {
    constructor() {
        this.id = '';
        this.pattern = new StepDefinitionPattern();
        this.sourceReference = new SourceReference();
    }
}
__decorate([
    class_transformer_1.Type(() => StepDefinitionPattern)
], StepDefinition.prototype, "pattern", void 0);
__decorate([
    class_transformer_1.Type(() => SourceReference)
], StepDefinition.prototype, "sourceReference", void 0);
exports.StepDefinition = StepDefinition;
class StepDefinitionPattern {
    constructor() {
        this.source = '';
        this.type = StepDefinitionPatternType.CUCUMBER_EXPRESSION;
    }
}
exports.StepDefinitionPattern = StepDefinitionPattern;
class TestCase {
    constructor() {
        this.id = '';
        this.pickleId = '';
        this.testSteps = [];
    }
}
__decorate([
    class_transformer_1.Type(() => TestStep)
], TestCase.prototype, "testSteps", void 0);
exports.TestCase = TestCase;
class Group {
    constructor() {
        this.children = [];
    }
}
__decorate([
    class_transformer_1.Type(() => Group)
], Group.prototype, "children", void 0);
exports.Group = Group;
class StepMatchArgument {
    constructor() {
        this.group = new Group();
    }
}
__decorate([
    class_transformer_1.Type(() => Group)
], StepMatchArgument.prototype, "group", void 0);
exports.StepMatchArgument = StepMatchArgument;
class StepMatchArgumentsList {
    constructor() {
        this.stepMatchArguments = [];
    }
}
__decorate([
    class_transformer_1.Type(() => StepMatchArgument)
], StepMatchArgumentsList.prototype, "stepMatchArguments", void 0);
exports.StepMatchArgumentsList = StepMatchArgumentsList;
class TestStep {
    constructor() {
        this.id = '';
    }
}
__decorate([
    class_transformer_1.Type(() => StepMatchArgumentsList)
], TestStep.prototype, "stepMatchArgumentsLists", void 0);
exports.TestStep = TestStep;
class TestCaseFinished {
    constructor() {
        this.testCaseStartedId = '';
        this.timestamp = new Timestamp();
    }
}
__decorate([
    class_transformer_1.Type(() => Timestamp)
], TestCaseFinished.prototype, "timestamp", void 0);
exports.TestCaseFinished = TestCaseFinished;
class TestCaseStarted {
    constructor() {
        this.attempt = 0;
        this.id = '';
        this.testCaseId = '';
        this.timestamp = new Timestamp();
    }
}
__decorate([
    class_transformer_1.Type(() => Timestamp)
], TestCaseStarted.prototype, "timestamp", void 0);
exports.TestCaseStarted = TestCaseStarted;
class TestRunFinished {
    constructor() {
        this.success = false;
        this.timestamp = new Timestamp();
    }
}
__decorate([
    class_transformer_1.Type(() => Timestamp)
], TestRunFinished.prototype, "timestamp", void 0);
exports.TestRunFinished = TestRunFinished;
class TestRunStarted {
    constructor() {
        this.timestamp = new Timestamp();
    }
}
__decorate([
    class_transformer_1.Type(() => Timestamp)
], TestRunStarted.prototype, "timestamp", void 0);
exports.TestRunStarted = TestRunStarted;
class TestStepFinished {
    constructor() {
        this.testCaseStartedId = '';
        this.testStepId = '';
        this.testStepResult = new TestStepResult();
        this.timestamp = new Timestamp();
    }
}
__decorate([
    class_transformer_1.Type(() => TestStepResult)
], TestStepFinished.prototype, "testStepResult", void 0);
__decorate([
    class_transformer_1.Type(() => Timestamp)
], TestStepFinished.prototype, "timestamp", void 0);
exports.TestStepFinished = TestStepFinished;
class TestStepResult {
    constructor() {
        this.duration = new Duration();
        this.status = TestStepResultStatus.UNKNOWN;
        this.willBeRetried = false;
    }
}
__decorate([
    class_transformer_1.Type(() => Duration)
], TestStepResult.prototype, "duration", void 0);
exports.TestStepResult = TestStepResult;
class TestStepStarted {
    constructor() {
        this.testCaseStartedId = '';
        this.testStepId = '';
        this.timestamp = new Timestamp();
    }
}
__decorate([
    class_transformer_1.Type(() => Timestamp)
], TestStepStarted.prototype, "timestamp", void 0);
exports.TestStepStarted = TestStepStarted;
class Timestamp {
    constructor() {
        this.seconds = 0;
        this.nanos = 0;
    }
}
exports.Timestamp = Timestamp;
class UndefinedParameterType {
    constructor() {
        this.expression = '';
        this.name = '';
    }
}
exports.UndefinedParameterType = UndefinedParameterType;
var AttachmentContentEncoding;
(function (AttachmentContentEncoding) {
    AttachmentContentEncoding["IDENTITY"] = "IDENTITY";
    AttachmentContentEncoding["BASE64"] = "BASE64";
})(AttachmentContentEncoding = exports.AttachmentContentEncoding || (exports.AttachmentContentEncoding = {}));
var SourceMediaType;
(function (SourceMediaType) {
    SourceMediaType["TEXT_X_CUCUMBER_GHERKIN_PLAIN"] = "text/x.cucumber.gherkin+plain";
    SourceMediaType["TEXT_X_CUCUMBER_GHERKIN_MARKDOWN"] = "text/x.cucumber.gherkin+markdown";
})(SourceMediaType = exports.SourceMediaType || (exports.SourceMediaType = {}));
var StepDefinitionPatternType;
(function (StepDefinitionPatternType) {
    StepDefinitionPatternType["CUCUMBER_EXPRESSION"] = "CUCUMBER_EXPRESSION";
    StepDefinitionPatternType["REGULAR_EXPRESSION"] = "REGULAR_EXPRESSION";
})(StepDefinitionPatternType = exports.StepDefinitionPatternType || (exports.StepDefinitionPatternType = {}));
var TestStepResultStatus;
(function (TestStepResultStatus) {
    TestStepResultStatus["UNKNOWN"] = "UNKNOWN";
    TestStepResultStatus["PASSED"] = "PASSED";
    TestStepResultStatus["SKIPPED"] = "SKIPPED";
    TestStepResultStatus["PENDING"] = "PENDING";
    TestStepResultStatus["UNDEFINED"] = "UNDEFINED";
    TestStepResultStatus["AMBIGUOUS"] = "AMBIGUOUS";
    TestStepResultStatus["FAILED"] = "FAILED";
})(TestStepResultStatus = exports.TestStepResultStatus || (exports.TestStepResultStatus = {}));
//# sourceMappingURL=messages.js.map