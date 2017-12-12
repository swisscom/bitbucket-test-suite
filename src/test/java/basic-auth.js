function() {
    var temp = karate.properties['username'] + ':' + karate.properties['password'];
    var Base64 = Java.type('java.util.Base64');
    var encoded = Base64.getEncoder().encodeToString(temp.bytes);
    return 'Basic ' + encoded;
}