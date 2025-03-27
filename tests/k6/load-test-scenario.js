import http from 'k6/http';
import { check, sleep } from 'k6';
import { generateUUID, getCurrentTime, getFutureTime, checkResponse, connectHeaders, endpoints } from './helpers/utils.js';

export const options = {
  scenarios: {
    issue_coupons: {
      executor: 'constant-arrival-rate',
      rate: parseInt(__ENV.REQUEST_RATE || '700'),  // 초당 요청 수 (기본값: 700 RPS)
      timeUnit: '1s',                                     // 1초당
      duration: __ENV.DURATION || '1m',                   // 테스트 지속 시간 (기본값: 5분)
      preAllocatedVUs: 100,                               // 미리 할당할 가상 유저 수
      maxVUs: 200,                                        // 최대 가상 유저 수
      exec: 'issueCoupon',
    }
  },
  thresholds: {
    http_req_duration: ['p(95)<500'],                            // 95%의 요청이 500ms 이내에 완료되어야 함
    'http_req_duration{scenario:issue_coupons}': ['p(95)<200'],  // 쿠폰 발급은 200ms 이내
    http_req_failed: ['rate<0.05'],                              // 에러율 5% 미만
  },
};

export function setup() {
  const host = __ENV.SERVER_HOST || 'localhost:8080';
  console.log(`서버 주소: http://${host}`);
  console.log(`CreateCampaign 엔드포인트: ${endpoints.createCampaign}`);

  const couponAmount = parseInt(__ENV.COUPON_AMOUNT || '10000'); // 발행할 쿠폰 갯수
  const now = new Date();
  const tomorrow = new Date(now.getTime() + 24 * 60 * 60 * 1000);
  
  const payload = JSON.stringify({
    name: `LoadTest`,
    amount: couponAmount,
    issued_at: now.toISOString(),
    expires_at: tomorrow.toISOString()
  });
  
  console.log('캠페인 생성 요청 페이로드:', payload);
  
  const url = `http://${host}${endpoints.createCampaign}`;
  console.log(`요청 URL: ${url}`);
  
  const res = http.post(url, payload, { headers: connectHeaders });
  
  console.log(`응답 상태 코드: ${res.status}`);

  if (res.status !== 200) {
    console.error('캠페인 생성 실패. 상태 코드:', res.status);
    console.error('응답 본문:', res.body);
    throw new Error(`Campaign creation failed with status ${res.status}`);
  }
  
  let campaignId = '';
  
  try {
    const body = JSON.parse(res.body);
    
    if (body.value && body.value.data && body.value.data.campaign) {
      campaignId = body.value.data.campaign.id;
    } else if (body.data && body.data.campaign) {
      campaignId = body.data.campaign.id;
    } else {
      console.log('응답 구조 탐색:');
      console.log(JSON.stringify(body, null, 2));
      
      const findId = (obj, path = '') => {
        if (typeof obj !== 'object' || obj === null) return null;
        
        if (obj.id && obj.name) {
          console.log(`ID 필드 발견: ${path}.id = ${obj.id}`);
          return obj.id;
        }
        
        for (const key in obj) {
          const result = findId(obj[key], `${path}.${key}`);
          if (result) return result;
        }
        
        return null;
      };
      
      campaignId = findId(body);
    }
    
    if (!campaignId) {
      console.error('응답에서 캠페인 ID를 찾을 수 없습니다. 응답 본문:', res.body);
      throw new Error('Campaign ID not found in response');
    }
    
    console.log(`생성된 캠페인 ID: ${campaignId}`);
    console.log(`쿠폰 수량: ${couponAmount}`);
  } catch (e) {
    console.error('응답 파싱 실패:', e);
    throw new Error(`Failed to parse response: ${e.message}`);
  }
  
  sleep(1);
  
  return {
    campaignId: campaignId,
    couponAmount: couponAmount
  };
}

export function issueCoupon(data) {
  if (!data || !data.campaignId) {
    console.error('캠페인 ID를 찾을 수 없습니다. 테스트를 중단합니다.');
    return;
  }
  
  const userId = generateUUID();
  
  const payload = JSON.stringify({
    campaign_id: data.campaignId,
    user_id: userId
  });
  
  const url = `http://${__ENV.SERVER_HOST || 'localhost:8080'}${endpoints.issueCoupon}`;
  const res = http.post(url, payload, { headers: connectHeaders });
  
  const validStatus = (r) => r.status === 200 || r.status === 400;
  check(res, {
    'Status is valid (200 or 400)': validStatus,
    'Response time is acceptable': (r) => r.timings.duration < 200
  });
}

export function teardown(data) {
  if (!data || !data.campaignId) {
    console.log('테스트 중 캠페인 ID를 찾을 수 없었습니다.');
    return;
  }
  
  console.log(`\n=============== 부하 테스트 완료 ===============`);
  console.log(`캠페인 ID: ${data.campaignId}`);
  console.log(`총 쿠폰 수량: ${data.couponAmount}`);
  console.log(`================================================\n`);
  
  try {
    const url = `http://${__ENV.SERVER_HOST || 'localhost:8080'}${endpoints.getCampaign}`;
    const payload = JSON.stringify({
      id: data.campaignId
    });
    
    const res = http.post(url, payload, { headers: connectHeaders });
    
    if (res.status === 200) {
      const body = JSON.parse(res.body);
      console.log('최종 캠페인 상태:');
      
      if (body.value && body.value.data && body.value.data.campaign) {
        const campaign = body.value.data.campaign;
        const issuedCount = campaign.issuedCoupons ? campaign.issuedCoupons.length : 0;
        console.log(`발급된 쿠폰 수: ${issuedCount}/${data.couponAmount} (${(issuedCount/data.couponAmount*100).toFixed(2)}%)`);
      } else {
        console.log('캠페인 정보를 응답에서 찾을 수 없습니다.');
      }
    } else {
      console.log(`캠페인 조회 실패. 상태 코드: ${res.status}`);
    }
  } catch (e) {
    console.error('최종 캠페인 상태 확인 실패:', e);
  }
}
