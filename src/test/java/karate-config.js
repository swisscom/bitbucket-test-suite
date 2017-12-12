function() {
  karate.configure('connectTimeout', 5000);
  karate.configure('readTimeout', 5000);
  if(karate.properties['proxy']) {
    karate.configure('proxy', karate.properties['proxy']);
  }
  return { demoBaseUrl: karate.properties['url'], project: karate.properties['project'] };
}
