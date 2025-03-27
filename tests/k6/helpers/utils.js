export function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    var r = Math.random() * 16 | 0,
        v = c == 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

export function getCurrentTime() {
  return new Date().toISOString();
}

export function getFutureTime(hours) {
  const date = new Date();
  date.setHours(date.getHours() + hours);
  return date.toISOString();
}

export function checkResponse(response, expectedStatus = 200) {
  const checks = {};
  
  checks[`Status is ${expectedStatus}`] = (r) => r.status === expectedStatus;
  checks['Response is valid JSON'] = (r) => {
    try {
      JSON.parse(r.body);
      return true;
    } catch (e) {
      return false;
    }
  };
  
  return checks;
}

export const connectHeaders = {
  'Content-Type': 'application/json',
  'Connect-Protocol-Version': '1',
  'Accept': 'application/json'
};

export const endpoints = {
  createCampaign: '/io.coupon.service.GreetService/CreateCampaign',
  issueCoupon: '/io.coupon.service.GreetService/IssueCoupon',
  getCampaign: '/io.coupon.service.GreetService/GetCampaign'
};
