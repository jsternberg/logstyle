# Logstyle

The `logstyle` command confirms if a package follows the InfluxDB [logging style guide](https://github.com/influxdata/influxdb/blob/master/logger/style_guide.md) by analyzing the locations where logging functions are used.

It performs the following checks:

* Verifies that the message argument is a constant string or a literal.
* Verifies the message uses proper capitalization and punctuation. (TODO)
* Verifies the field keys are in `snake_case`. (TODO)
* Verifies that if an error is used in logging, the error is not returned from the function. (TODO)
