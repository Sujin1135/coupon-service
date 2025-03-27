# K6 부하 테스트 가이드

이 디렉토리에는 [k6](https://k6.io/)를 사용하여 쿠폰 서비스의 부하 테스트를 수행하는 스크립트가 포함되어 있습니다.

## 준비 사항

### K6 설치

#### macOS:
```bash
brew install k6
```

## 테스트 실행 방법

### 쿠폰발급 부하 테스트
**부하 테스트 전에 서버가 실행중인 상태여야 합니다!**
```bash
k6 run load-test-scenario.js
```

## 테스트 매개변수 설정

다음과 같은 환경 변수를 사용하여 테스트 매개변수를 설정할 수 있습니다:

- **SERVER_HOST**: 서버 호스트 (기본값: localhost:8080)
- **REQUEST_RATE**: 초당 요청 수 (기본값: 700)
- **DURATION**: 테스트 지속 시간 (기본값: 1m)
- **COUPON_AMOUNT**: 생성할 쿠폰 수량 (기본값: 10000)

예시:
```bash
k6 run -e SERVER_HOST=localhost:8080 -e REQUEST_RATE=700 -e DURATION=1m -e COUPON_AMOUNT=10000 load-test-scenario.js
```

## 결과 분석

k6는 테스트 결과를 다양한 형식으로 출력할 수 있습니다. JSON 형식으로 결과를 저장하려면:

```bash
k6 run --out json=results.json load-test-scenario.js
```

## 로컬 테스트 수행 예시

간단히 Mac OS 에서 아래와 같은 스팩을 지닌 로컬 환경에서 싱글 인스턴스를 띄워 테스트를 수행 하였고 아래 이미지와 같은 결과를 얻었습니다. <br/>
만약 적당한 스팩에 여러 인스턴스가 라우팅된 환경과 적절한 DB 스팩에서 테스트 한다면 더 높은 부하까지 견딜 수 있을것으로 생각됩니다.

### 스팩
- Chip: Apple M2
- Memory: 16GB

![Screenshot 2025-03-27 at 8.12.14 PM.png](../../public/images/Screenshot%202025-03-27%20at%208.12.14%E2%80%AFPM.png)
