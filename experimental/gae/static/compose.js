if (!window.lightwave) {
    lightwave = {};
}

if (!lightwave.ot) {
    lightwave.ot = {};
}

// ------------------------------------------------------------------

lightwave.ot.composeReader = function(stream1, stream2) {
    this.stream1 = stream1;
    this.stream2 = stream2;
};

// Read a tuple of operations from stream1 and stream2
lightwave.ot.composeReader.prototype.Read = function() {
    var second, first, err;
    // EOF?
    if (this.stream1.IsEOF() && this.stream2.IsEOF()) {
        return [null, null, null];
    }
    // EOF Stream1?
    if (this.stream1.IsEOF()) {
        if (!this.stream2.ops[this.stream2.pos]["i"]) {
            err = lightwave.NewError("Streams have different length (1)")
            return [null, null, err];
        }
        first = this.stream2.Read(-1)
        return [first, second, null];
    }
    // EOF Stream2?
    if (this.stream2.IsEOF()) {
        if (!this.stream1.ops[this.stream1.pos]["i"]) {
            err = os.NewError("Streams have different length (2)")
            return [null, null, err];
        }
        second = this.stream1.Read(-1)
        return [first, second, null];
    }
    // Insert of stream1 goes first
    if (this.stream1.ops[this.stream1.pos]["i"]) {
        second = this.stream1.Read(-1)
        return [first, second, null];
    }
    op1_len = lightwave.ot.opLen(this.stream1.ops[this.stream1.pos]);
    op2_len = lightwave.ot.opLen(this.stream2.ops[this.stream2.pos]);
    // Skip, Insert (of stream2) and Delete go together
    var l = Math.min(op1_len - this.stream1.inside, op2_len - this.stream2.inside);
    first = this.stream1.Read( l );
    second = this.stream2.Read( l );
    return [first, second, null];
};

lightwave.ot.composeStringOperations = function(first, second, f) {
    var result = [], err;
    var reader = new lightwave.ot.composeReader(new lightwave.ot.Stream(second), new lightwave.ot.Stream(first));
    while (true) {
        var first_op, second_op;
        var tmp = reader.Read()
        second_op = tmp[0];
        first_op = tmp[1];
        err = tmp[2];
        // Error or EOF?
        if (err || (!first_op && !second_op)) {
            return [result, err];
        }
        var op;
        var tmp = lightwave.ot.composeStringOperation(first_op, second_op);
        op = tmp[0];
        err = tmp[1];
        if (err) {
            return [null, err];
        }
        if (op) {
            result.push(op);
        }
    }
    return [result, err];
};

lightwave.ot.composeStringOperation = function(first, second) {
    var result, err;
    if (first && !first["i"] && !first["s"] && !first["d"] && !first["t"]) {
        err = os.NewError("Operation not allowed in a string: first");
        return [result, err];
    }
    if (second && !second["i"] && !second["s"] && !second["d"] && !second["t"]) {
        err = os.NewError("Operation not allowed in a string: second");
        return [result, err];
    }
    if (first && (first["i"] || first["t"])) {
        if (second && second["d"]) {
            result = {"t": lightwave.ot.opLen(first)}; // Insert a tomb in the composed op
        } else {
            result = first;
        }
    } else if (first && first["d"]) {
        result = {"d": lightwave.ot.opLen(first)};
    } else {
        result = second;
    }
    return [result, err];
};
